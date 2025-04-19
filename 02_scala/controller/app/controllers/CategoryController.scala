package controllers

import play.api.mvc._
import play.api.libs.json._
import javax.inject._
import models.Category

@Singleton
class CategoryController @Inject()(cc: ControllerComponents) extends AbstractController(cc) {

  // Lista kategorii w pamięci
  private var categories = List(
    Category(1, "Electronics"),
    Category(2, "Books"),
    Category(3, "Clothing")
  )

  // Pobierz wszystkie kategorie
  def getAll: Action[AnyContent] = Action {
    Ok(Json.toJson(categories))
  }

  // Pobierz kategorię po id
  def getById(id: Long): Action[AnyContent] = Action {
    categories.find(_.id == id) match {
      case Some(category) => Ok(Json.toJson(category))
      case None           => NotFound(Json.obj("error" -> "Category not found"))
    }
  }

  // Dodaj nową kategorię
  def addCategory: Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[Category].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid category format")),
      category => {
         val newId = if (categories.isEmpty) 1 else categories.map(_.id).max + 1
         val newCategory = category.copy(id = newId)
         categories = categories :+ newCategory
         Created(Json.toJson(newCategory))
      }
    )
  }

  // Aktualizuj kategorię
  def updateCategory(id: Long): Action[JsValue] = Action(parse.json) { request =>
    request.body.validate[Category].fold(
      errors => BadRequest(Json.obj("error" -> "Invalid category format")),
      updatedCategory => {
        if(categories.exists(_.id == id)) {
          categories = categories.map { c =>
            if(c.id == id) updatedCategory.copy(id = id) else c
          }
          Ok(Json.toJson(updatedCategory.copy(id = id)))
        } else {
          NotFound(Json.obj("error" -> "Category not found"))
        }
      }
    )
  }

  // Usuń kategorię
  def deleteCategory(id: Long): Action[AnyContent] = Action {
    if(categories.exists(_.id == id)) {
      categories = categories.filterNot(_.id == id)
      Ok(Json.obj("status" -> "Category deleted"))
    } else {
      NotFound(Json.obj("error" -> "Category not found"))
    }
  }
}
