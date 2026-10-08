package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// indexView agrupa los datos que necesita la plantilla tasks/index.plush.html.
type indexView struct {
	// task y errors alimentan el formulario de creación.
	task   *models.Task
	errors *validate.Errors
	// editTask y editErrors alimentan el formulario de edición en línea.
	// editTask es nil cuando no se está editando ninguna tarea.
	editTask   *models.Task
	editErrors *validate.Errors
}

// renderIndex carga la lista de tareas y renderiza la página principal.
func renderIndex(c buffalo.Context, tx *pop.Connection, status int, v indexView) error {
	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
		return err
	}

	if v.task == nil {
		v.task = &models.Task{}
	}
	if v.errors == nil {
		v.errors = validate.NewErrors()
	}
	if v.editErrors == nil {
		v.editErrors = validate.NewErrors()
	}

	editingID := ""
	editTask := &models.Task{}
	if v.editTask != nil {
		editingID = v.editTask.ID.String()
		editTask = v.editTask
	}

	c.Set("tasks", tasks)
	c.Set("task", v.task)
	c.Set("errors", v.errors)
	c.Set("editingID", editingID)
	c.Set("editTask", editTask)
	c.Set("editErrors", v.editErrors)
	return c.Render(status, r.HTML("tasks/index.plush.html"))
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	return renderIndex(c, tx, http.StatusOK, indexView{})
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderIndex(c, tx, http.StatusUnprocessableEntity, indexView{task: task, errors: verrs})
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra la lista con el formulario de edición abierto
// para la tarea indicada.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return renderIndex(c, tx, http.StatusOK, indexView{editTask: task})
}

// TasksUpdate cambia el título de una tarea existente.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Solo el título se acepta desde el formulario (el resto de campos
	// del modelo tienen form:"-").
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderIndex(c, tx, http.StatusUnprocessableEntity, indexView{editTask: task, editErrors: verrs})
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
