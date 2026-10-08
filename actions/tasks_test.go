package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) createTask(title string, completed bool) *models.Task {
	task := &models.Task{Title: title, Completed: completed}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) Test_TasksIndex_ShowsEditLink() {
	task := as.createTask("Tarea con enlace de edición", false)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
}

func (as *ActionSuite) Test_TasksEdit_ShowsForm() {
	task := as.createTask("Título original", false)
	other := as.createTask("Otra tarea", false)

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `action="/tasks/`+task.ID.String()+`"`)
	as.Contains(body, `value="Título original"`)
	as.Contains(body, "Guardar")
	as.Contains(body, "Cancelar")
	// La otra tarea sigue mostrándose en modo lectura.
	as.Contains(body, "/tasks/"+other.ID.String()+"/edit")
	as.NotContains(body, `action="/tasks/`+other.ID.String()+`" method="POST" class="edit-task"`)
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/edit", id).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_Valid() {
	task := as.createTask("Título original", true)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"  Título corregido  "}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	as.NoError(as.DB.Reload(task))
	as.Equal("Título corregido", task.Title)
	// Editar el título no cambia el estado de la tarea.
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_OnlyChangesTitle() {
	task := as.createTask("Título original", false)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{
		"title":     {"Título nuevo"},
		"completed": {"true"},
		"id":        {uuid.Must(uuid.NewV4()).String()},
	})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.Equal("Título nuevo", task.Title)
	as.False(task.Completed)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle_Invalid() {
	task := as.createTask("Título original", false)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces_Invalid() {
	task := as.createTask("Título original", false)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"     "}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TooLong_Invalid() {
	task := as.createTask("Título original", false)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {strings.Repeat("a", models.TitleMaxLength+1)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Put(url.Values{"title": {"Lo que sea"}})
	as.Equal(404, res.Code)
}
