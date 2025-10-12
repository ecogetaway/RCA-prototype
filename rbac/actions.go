package rbac

import (
	"github.com/coroot/coroot/model"
)

var (
	Actions ActionSet
)

type ActionSet struct{}

func (as ActionSet) Settings() Settings {
	return Settings{}
}

func (as ActionSet) Users() UsersActionSet {
	return UsersActionSet{}
}

func (as ActionSet) Roles() RolesActionSet {
	return RolesActionSet{}
}

func (as ActionSet) Project(id string) ProjectActionSet {
	return ProjectActionSet{id: id}
}

func (as ActionSet) List() []Action {
	return append([]Action{
		as.Settings().Edit(),
		as.Users().Edit(),
		as.Roles().Edit(),
	},
		as.Project("").List()...)
}

type Settings struct{}

func (as Settings) Edit() Action {
	return NewAction(ScopeSettings, ActionEdit, nil)
}

type UsersActionSet struct{}

func (as UsersActionSet) Edit() Action {
	return NewAction(ScopeUsers, ActionEdit, nil)
}

type RolesActionSet struct{}

func (as RolesActionSet) Edit() Action {
	return NewAction(ScopeRoles, ActionEdit, nil)
}

type ProjectActionSet struct {
	id string
}

func (as ProjectActionSet) object() Object {
	return Object{"project_id": as.id}
}

func (as ProjectActionSet) List() []Action {
	return []Action{
		as.Settings().Edit(),
		as.Integrations().Edit(),
		as.ApplicationCategories().Edit(),
		as.CustomApplications().Edit(),
		as.CustomCloudPricing().Edit(),
		as.Inspections().Edit(),
		as.Instrumentations().Edit(),
		as.Traces().View(),
		as.Logs().View(),
		as.Costs().View(),
		as.Anomalies().View(),
		as.Risks().View(),
		as.Risks().Edit(),
		as.Application("*", "*", "*", "*").View(),
		as.Node("*").View(),
		as.Dashboards().Edit(),
		as.Dashboard("*").View(),
	}
}

func (as ProjectActionSet) Settings() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectSettings}
}

func (as ProjectActionSet) Integrations() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectIntegrations}
}

func (as ProjectActionSet) ApplicationCategories() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectApplicationCategories}
}

func (as ProjectActionSet) CustomApplications() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectCustomApplications}
}

func (as ProjectActionSet) CustomCloudPricing() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectCustomCloudPricing}
}

func (as ProjectActionSet) Inspections() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectInspections}
}

func (as ProjectActionSet) Instrumentations() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeProjectInstrumentations}
}

func (as ProjectActionSet) Traces() ProjectViewAction {
	return ProjectViewAction{project: &as, scope: ScopeProjectTraces}
}

func (as ProjectActionSet) Logs() ProjectViewAction {
	return ProjectViewAction{project: &as, scope: ScopeProjectLogs}
}

func (as ProjectActionSet) Costs() ProjectViewAction {
	return ProjectViewAction{project: &as, scope: ScopeProjectCosts}
}

func (as ProjectActionSet) Anomalies() ProjectViewAction {
	return ProjectViewAction{project: &as, scope: ScopeProjectAnomalies}
}

func (as ProjectActionSet) Risks() ProjectAction {
	return ProjectAction{project: &as, scope: ScopeProjectRisks}
}

func (as ProjectActionSet) Application(category model.ApplicationCategory, namespace string, kind model.ApplicationKind, name string) ApplicationActionSet {
	return ApplicationActionSet{project: &as, category: category, namespace: namespace, kind: kind, name: name}
}

func (as ProjectActionSet) Node(name string) NodeActionSet {
	return NodeActionSet{project: &as, name: name}
}

func (as ProjectActionSet) Dashboards() ProjectEditAction {
	return ProjectEditAction{project: &as, scope: ScopeDashboards}
}

func (as ProjectActionSet) Dashboard(name string) DashboardActionSet {
	return DashboardActionSet{project: &as, name: name}
}

type ProjectViewAction struct {
	project *ProjectActionSet
	scope   Scope
}

func (as ProjectViewAction) View() Action {
	return NewAction(as.scope, ActionView, as.project.object())
}

type ProjectEditAction struct {
	project *ProjectActionSet
	scope   Scope
}

func (as ProjectEditAction) Edit() Action {
	return NewAction(as.scope, ActionEdit, as.project.object())
}

type ProjectAction struct {
	project *ProjectActionSet
	scope   Scope
}

func (as ProjectAction) View() Action {
	return NewAction(as.scope, ActionView, as.project.object())
}

func (as ProjectAction) Edit() Action {
	return NewAction(as.scope, ActionEdit, as.project.object())
}

type ApplicationActionSet struct {
	project   *ProjectActionSet
	category  model.ApplicationCategory
	namespace string
	kind      model.ApplicationKind
	name      string
}

func (as ApplicationActionSet) object() Object {
	o := as.project.object()
	o["application_category"] = string(as.category)
	o["application_namespace"] = as.namespace
	o["application_kind"] = string(as.kind)
	o["application_name"] = as.name
	return o
}

func (as ApplicationActionSet) View() Action {
	return NewAction(ScopeApplication, ActionView, as.object())
}

type NodeActionSet struct {
	project *ProjectActionSet
	name    string
}

func (as NodeActionSet) object() Object {
	o := as.project.object()
	o["node_name"] = as.name
	return o
}

func (as NodeActionSet) View() Action {
	return NewAction(ScopeNode, ActionView, as.object())
}

type DashboardActionSet struct {
	project *ProjectActionSet
	name    string
}

func (as DashboardActionSet) object() Object {
	o := as.project.object()
	o["dashboard_name"] = as.name
	return o
}

func (as DashboardActionSet) View() Action {
	return NewAction(ScopeDashboard, ActionView, as.object())
}
