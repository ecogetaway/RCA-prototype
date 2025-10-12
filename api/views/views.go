package views

import (
	"context"
	"net/url"

	"github.com/coroot/coroot/api/views/application"
	"github.com/coroot/coroot/api/views/applications"
	"github.com/coroot/coroot/api/views/aws"
	"github.com/coroot/coroot/api/views/dashboards"
	"github.com/coroot/coroot/api/views/incident"
	"github.com/coroot/coroot/api/views/inspections"
	"github.com/coroot/coroot/api/views/logs"
	"github.com/coroot/coroot/api/views/overview"
	"github.com/coroot/coroot/api/views/profiling"
	"github.com/coroot/coroot/api/views/roles"
	"github.com/coroot/coroot/api/views/tracing"
	"github.com/coroot/coroot/api/views/users"
	"github.com/coroot/coroot/clickhouse"
	"github.com/coroot/coroot/db"
	"github.com/coroot/coroot/model"
	"github.com/coroot/coroot/rbac"
)

func Overview(ctx context.Context, ch *clickhouse.Client, p *db.Project, w *model.World, view, query string) *overview.Overview {
	return overview.Render(ctx, ch, p, w, view, query)
}

func Application(p *db.Project, w *model.World, app *model.Application) *application.View {
	return application.Render(p, w, app)
}

func Incident(w *model.World, app *model.Application, i *model.ApplicationIncident) *incident.View {
	return incident.Render(w, app, i)
}

func Incidents(w *model.World, incidents []*model.ApplicationIncident) []incident.Incident {
	return incident.RenderList(w, incidents)
}

func Profiling(ctx context.Context, ch *clickhouse.Client, app *model.Application, q url.Values, w *model.World) *profiling.View {
	return profiling.Render(ctx, ch, app, q, w)
}

func Tracing(ctx context.Context, ch *clickhouse.Client, app *model.Application, q url.Values, w *model.World) *tracing.View {
	return tracing.Render(ctx, ch, app, q, w)
}

func Logs(ctx context.Context, ch *clickhouse.Client, app *model.Application, q url.Values, w *model.World) *logs.View {
	return logs.Render(ctx, ch, app, q, w)
}

func Inspections(checkConfigs model.CheckConfigs) *inspections.View {
	return inspections.Render(checkConfigs)
}

func CustomApplications(p *db.Project) *applications.CustomApplicationsView {
	return applications.RenderCustomApplications(p)
}

func AWS(w *model.World) *aws.View {
	return aws.Render(w)
}

func Roles(rs []rbac.Role) *roles.View {
	return roles.Render(rs)
}

func Users(us []*db.User, rs []rbac.Role) *users.Users {
	return users.RenderUsers(us, rs)
}

func User(u *db.User, projects map[db.ProjectId]string, viewonly bool) *users.User {
	return users.RenderUser(u, projects, viewonly)
}

var (
	Dashboards = &dashboards.Dashboards{}
)
