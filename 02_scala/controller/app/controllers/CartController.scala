package controllers

import play.api.mvc._
import play.api.libs.json._
import javax.inject._
import models.CartItem

@Singleton
class CartController @Inject()(cc: ControllerComponents) extends AbstractController(cc) {

  // Lista pozycji w koszyku w pamięci
  private var cartItems = List(
    CartItem(1, productId = 1, quantity = 2), // np. 2 sztuki produktu o id 1
    CartItem(2, productId = 3, quantity = 1)
  )

  // Pobierz wszystkie elementy koszyka
  def getAll: Action[AnyContent] = Action {
    Ok(Json.toJson(cartItems))
  }

  // Pobierz element koszyka po id
  def getById(id: Long): Action[AnyContent] = Action {
    cartItems.find(_.id == id) match {
      case Some(item) => Ok(Json.toJson(item))
      case None       => NotFound(Json.obj("error" -> "Cart item not found"))
    }
  }

  // Dodaj nowy element do koszyka
  def addItem: Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[CartItem].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid cart item format")),
      cartItem => {
         val newId = if (cartItems.isEmpty) 1 else cartItems.map(_.id).max + 1
         val newItem = cartItem.copy(id = newId)
         cartItems = cartItems :+ newItem
         Created(Json.toJson(newItem))
      }
    )
  }

  // Aktualizuj element koszyka
  def updateItem(id: Long): Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[CartItem].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid cart item format")),
      updatedItem => {
        if(cartItems.exists(_.id == id)) {
          cartItems = cartItems.map { item =>
            if(item.id == id) updatedItem.copy(id = id) else item
          }
          Ok(Json.toJson(updatedItem.copy(id = id)))
        } else {
          NotFound(Json.obj("error" -> "Cart item not found"))
        }
      }
    )
  }

  // Usuń element z koszyka
  def deleteItem(id: Long): Action[AnyContent] = Action {
    if(cartItems.exists(_.id == id)) {
      cartItems = cartItems.filterNot(_.id == id)
      Ok(Json.obj("status" -> "Cart item deleted"))
    } else {
      NotFound(Json.obj("error" -> "Cart item not found"))
    }
  }
}
