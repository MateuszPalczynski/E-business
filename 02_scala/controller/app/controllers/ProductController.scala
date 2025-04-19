package controllers

import play.api.mvc._
import play.api.libs.json._
import javax.inject._
import models.Product

@Singleton
class ProductController @Inject()(cc: ControllerComponents) extends AbstractController(cc) {

  // Lista produktów w pamięci
  private var products = List(
    Product(1, "Laptop", 3999.99),
    Product(2, "Smartfon", 2499.50),
    Product(3, "Monitor", 899.00)
  )

  // Pobierz wszystkie produkty
  def getAll: Action[AnyContent] = Action {
    Ok(Json.toJson(products))
  }

  // Pobierz produkt po id
  def getById(id: Long): Action[AnyContent] = Action {
    products.find(_.id == id) match {
      case Some(product) => Ok(Json.toJson(product))
      case None          => NotFound(Json.obj("error" -> "Product not found"))
    }
  }

  // Dodaj nowy produkt
  def addProduct: Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[Product].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid product format")),
      product => {
         val newId = if (products.isEmpty) 1 else products.map(_.id).max + 1
         val newProduct = product.copy(id = newId)
         products = products :+ newProduct
         Created(Json.toJson(newProduct))
      }
    )
  }

  // Aktualizuj istniejący produkt
  def updateProduct(id: Long): Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[Product].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid product format")),
      updatedProduct => {
        if(products.exists(_.id == id)) {
          products = products.map { p =>
            if(p.id == id) updatedProduct.copy(id = id) else p
          }
          Ok(Json.toJson(updatedProduct.copy(id = id)))
        } else {
          NotFound(Json.obj("error" -> "Product not found"))
        }
      }
    )
  }

  // Usuń produkt
  def deleteProduct(id: Long): Action[AnyContent] = Action {
    if(products.exists(_.id == id)) {
      products = products.filterNot(_.id == id)
      Ok(Json.obj("status" -> "Product deleted"))
    } else {
      NotFound(Json.obj("error" -> "Product not found"))
    }
  }
}
