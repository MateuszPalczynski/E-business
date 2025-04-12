package controllers

import javax.inject._
import play.api.mvc._
import play.api.libs.json.{Json, OFormat, JsValue, Writes}

// Definicja modelu
case class Product(id: Int, name: String, price: Double)

// Obiekt JSON – UWAGA: jawnie podajemy typ OFormat[Product]
object Product {
  implicit val productFormat: OFormat[Product] = Json.format[Product]
  // Definicja formatera dla listy produktów, aby serializacja List[Product] działała poprawnie
  implicit val productListWrites: Writes[List[Product]] = Writes.list(productFormat)
}

@Singleton
class ProductController @Inject()(val controllerComponents: ControllerComponents) extends BaseController {

  // Lista produktów – symulacja bazy danych
  private var products: List[Product] = List(
    Product(1, "Produkt 1", 100.0),
    Product(2, "Produkt 2", 150.0)
  )

  // Endpoint GET /products – zwraca wszystkie produkty
  def getAll: Action[AnyContent] = Action {
    Ok(Json.toJson(products))
  }

  // Endpoint GET /products/:id – zwraca produkt po id
  def getById(id: Int): Action[AnyContent] = Action {
    products.find(_.id == id) match {
      case Some(product) => Ok(Json.toJson(product))
      case None => NotFound("Produkt nie znaleziony")
    }
  }

  // Endpoint POST /products – tworzy nowy produkt
  // Używamy Action(parse.json) z jawnie zadeklarowanym typem akcji (Action[JsValue])
  def create: Action[JsValue] = Action(parse.json) { (request: Request[JsValue]) =>
  request.body.validate[Product].fold(
    errors => BadRequest("Błędny format"),
    product => {
      products = products :+ product
      Created(Json.toJson(product))
    }
  )
}



  // Endpoint PUT /products/:id – aktualizuje produkt
  def update(id: Int): Action[JsValue] = Action(parse.json) { (request: Request[JsValue]) =>
  request.body.validate[Product].fold(
    errors => BadRequest("Błędny format"),
    updatedProduct => {
      products = products.map(p => if (p.id == id) updatedProduct else p)
      Ok(Json.toJson(updatedProduct))
    }
  )
}


  // Endpoint DELETE /products/:id – usuwa produkt
  def delete(id: Int): Action[AnyContent] = Action {
    val initialSize = products.size
    products = products.filterNot(_.id == id)
    if (products.size < initialSize) Ok("Usunięto produkt")
    else NotFound("Produkt nie znaleziony")
  }
}
