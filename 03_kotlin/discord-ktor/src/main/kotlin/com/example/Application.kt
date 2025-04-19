package com.example

import io.ktor.server.application.*
import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.ktor.server.routing.*
import io.ktor.server.response.*
import io.ktor.server.request.*
import io.ktor.http.*

import io.ktor.client.*
import io.ktor.client.engine.cio.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.plugins.websocket.*
import io.ktor.client.request.*
import io.ktor.client.statement.*

import io.ktor.serialization.kotlinx.json.*
import io.ktor.websocket.*

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.*

import kotlinx.coroutines.*
import java.time.Duration

// ─── MODELE DANYCH ────────────────────────────────────────────────────────────
@Serializable
data class DiscordMessage(val content: String)

@Serializable
data class Product(val id: Int, val name: String, val price: Double, val category: String)

// ─── PRZYKŁADOWE DANE ─────────────────────────────────────────────────────────
val sampleProducts = listOf(
    Product(1, "Laptop Gaming", 4999.99, "Electronics"),
    Product(2, "Smartphone", 2499.50, "Electronics"),
    Product(3, "Design T-Shirt", 89.99, "Clothing"),
    Product(4, "Jeans", 199.99, "Clothing"),
    Product(5, "Programming Book", 129.99, "Books"),
    Product(6, "Headphones", 349.99, "Electronics"),
    Product(7, "Running Shoes", 299.99, "Sports"),
    Product(8, "Coffee Maker", 599.99, "Home")
)

// ─── ENTRY POINT ──────────────────────────────────────────────────────────────
fun main(args: Array<String>): Unit = EngineMain.main(args)

// ─── MODUŁ APLIKACJI ──────────────────────────────────────────────────────────
@Suppress("unused")
fun Application.module() {
    println("🚀 Starting Discord Bot Application")

    val botToken = environment.config.property("ktor.discord.botToken").getString()
    val channelId = environment.config.property("ktor.discord.channelId").getString()

    // HTTP client
    val httpClient = HttpClient(CIO) {
        install(ContentNegotiation) { json() }
    }

    // WebSocket client
    val wsClient = HttpClient(CIO) {
        install(WebSockets)
        install(ContentNegotiation) { json() }
    }

    environment.monitor.subscribe(ApplicationStarted) {
        launch {
            startDiscordGatewayListener(wsClient, httpClient, botToken)
        }
    }

    routing {
        get("/") {
            call.respondText("Discord Bot is running ✅", ContentType.Text.Plain)
        }

        get("/send") {
            val msg = call.request.queryParameters["message"] ?: "Hello from Ktor!"
            val response = sendDiscordMessage(httpClient, botToken, channelId, msg)
            call.respondText(
                "Discord response: ${response.status}",
                ContentType.Text.Plain
            )
        }
    }
}

// ─── OBSŁUGA DISCORD GATEWAY ──────────────────────────────────────────────────
suspend fun startDiscordGatewayListener(
    client: HttpClient,
    httpClient: HttpClient,
    token: String
) {
    client.webSocket(
        request = {
            url {
                protocol = URLProtocol.WSS
                host = "gateway.discord.gg"
                encodedPath = "/"
                parameters.append("v", "10")
                parameters.append("encoding", "json")
            }
        }
    ) {
        // 1. Odbierz HELLO
        val helloText = (incoming.receive() as Frame.Text).readText()
        val helloJson = Json.parseToJsonElement(helloText).jsonObject
        val interval = helloJson["d"]!!
            .jsonObject["heartbeat_interval"]!!.jsonPrimitive.int

        // 2. Wyślij IDENTIFY
        val identify = buildJsonObject {
            put("op", 2)
            put("d", buildJsonObject {
                put("token", token)
                put("intents", (1 shl 0) or (1 shl 9) or (1 shl 15)) // GUILDS + GUILD_MESSAGES + MESSAGE_CONTENT
                put("properties", buildJsonObject {
                    put("\$os", "linux")
                    put("\$browser", "ktor")
                    put("\$device", "ktor")
                })
            })
        }
        send(Frame.Text(identify.toString()))

        // 3. Utrzymanie połączenia
        launch {
            while (true) {
                delay(interval.toLong())
                send(Frame.Text("""{"op":1,"d":null}"""))
            }
        }

        // 4. Obsługa zdarzeń
        for (frame in incoming) {
            if (frame is Frame.Text) {
                handleDiscordEvent(
                    frame.readText(),
                    httpClient,
                    token
                )
            }
        }
    }
}

// ─── OBSŁUGA ZDARZEŃ DISCORD ─────────────────────────────────────────────────
suspend fun handleDiscordEvent(
    eventJson: String,
    httpClient: HttpClient,
    token: String
) {
    val msg = Json.parseToJsonElement(eventJson).jsonObject

    when (msg["t"]?.jsonPrimitive?.contentOrNull) {
        "MESSAGE_CREATE" -> {
            val data = msg["d"]!!.jsonObject
            val content = data["content"]!!.jsonPrimitive.content
            val author = data["author"]!!.jsonObject["username"]!!.jsonPrimitive.content
            val channelId = data["channel_id"]!!.jsonPrimitive.content

            println("📥 Message from $author: $content")

            when {
                content == "!help" -> {
                    sendDiscordMessage(
                        httpClient,
                        token,
                        channelId,
                        """
                        **Available Commands:**
                        - `!categories` - List all categories
                        - `!products <category>` - List products in category
                        - `!help` - Show this help
                        """.trimIndent()
                    )
                }

                content == "!categories" -> {
                    handleCategoriesCommand(httpClient, token, channelId)
                }

                content.startsWith("!products") -> {
                    handleProductsCommand(content, httpClient, token, channelId)
                }
            }
        }
    }
}

// ─── OBSŁUGA KOMEND ──────────────────────────────────────────────────────────
suspend fun handleCategoriesCommand(
    httpClient: HttpClient,
    token: String,
    channelId: String
) {
    try {
        val categories = sampleProducts
            .map { it.category }
            .distinct()
            .sorted()

        val message = if (categories.isNotEmpty()) {
            "📂 **Available Categories:**\n" + categories.joinToString("\n") { "▫ $it" }
        } else {
            "❌ No categories available"
        }

        sendDiscordMessage(httpClient, token, channelId, message)
    } catch (e: Exception) {
        e.printStackTrace()
        sendDiscordMessage(httpClient, token, channelId, "❌ Error fetching categories")
    }
}

suspend fun handleProductsCommand(
    command: String,
    httpClient: HttpClient,
    token: String,
    channelId: String
) {
    try {
        val category = command.substringAfter("!products").trim()

        if (category.isBlank()) {
            sendDiscordMessage(
                httpClient,
                token,
                channelId,
                "❌ Please specify a category! Example: `!products Electronics`"
            )
            return
        }

        val products = sampleProducts
            .filter { it.category.equals(category, ignoreCase = true) }
            .sortedBy { it.name }

        val message = if (products.isNotEmpty()) {
            "🛍️ **Products in '$category':**\n" +
                    products.joinToString("\n") {
                        "▫ ${it.name} - ${it.price} PLN"
                    }
        } else {
            "❌ No products found in category '$category'"
        }

        sendDiscordMessage(httpClient, token, channelId, message)
    } catch (e: Exception) {
        e.printStackTrace()
        sendDiscordMessage(httpClient, token, channelId, "❌ Error processing products request")
    }
}

// ─── WYSYŁANIE WIADOMOŚCI ────────────────────────────────────────────────────
suspend fun sendDiscordMessage(
    client: HttpClient,
    token: String,
    channelId: String,
    content: String
): HttpResponse {
    return client.post("https://discord.com/api/v10/channels/$channelId/messages") {
        header("Authorization", "Bot $token")
        contentType(ContentType.Application.Json)
        setBody(DiscordMessage(content))
    }
}