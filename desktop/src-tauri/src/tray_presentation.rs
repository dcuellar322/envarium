use crate::model::DesktopSnapshot;

// Only native menu/tooltip inputs belong here. Snapshot polling and notification
// tracking continue even when the rendered menu has not changed.
#[derive(Debug, PartialEq, Eq)]
pub(crate) struct Presentation {
    status: String,
    version: String,
    project_count: usize,
    projects: Vec<(String, String, String)>,
    workspace_count: usize,
    workspaces: Vec<(String, String)>,
    keep_running: bool,
    autostart: bool,
}

impl Presentation {
    pub(crate) fn new(snapshot: &DesktopSnapshot, keep_running: bool, autostart: bool) -> Self {
        Self {
            status: snapshot.system.status.clone(),
            version: snapshot.system.version.clone(),
            project_count: snapshot.projects.len(),
            projects: snapshot
                .projects
                .iter()
                .take(crate::tray::RESOURCE_LIMIT)
                .map(|item| {
                    (
                        item.project.id.clone(),
                        item.project.display_name.clone(),
                        item.runtime
                            .as_ref()
                            .map(|runtime| runtime.state.clone())
                            .unwrap_or_else(|| "not observed".into()),
                    )
                })
                .collect(),
            workspace_count: snapshot.workspaces.len(),
            workspaces: snapshot
                .workspaces
                .iter()
                .take(crate::tray::RESOURCE_LIMIT)
                .map(|item| (item.id.clone(), item.name.clone()))
                .collect(),
            keep_running,
            autostart,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn snapshot() -> DesktopSnapshot {
        serde_json::from_str(
            r#"{
            "system":{"apiVersion":"v1","databaseSchemaVersion":18,"status":"ready","version":"1"},
            "projects":[{"project":{"id":"one","displayName":"One"},"runtime":{"state":"running"}}],
            "workspaces":[],"operations":[],"portConflictCount":0
        }"#,
        )
        .unwrap()
    }

    #[test]
    fn refreshes_changed_menu_inputs_but_ignores_unrendered_metrics() {
        let mut current = snapshot();
        let first = Presentation::new(&current, true, false);
        current.port_conflict_count = 2;
        assert_eq!(first, Presentation::new(&current, true, false));
        current.projects[0].runtime.as_mut().unwrap().state = "stopped".into();
        assert_ne!(first, Presentation::new(&current, true, false));
        assert_ne!(first, Presentation::new(&snapshot(), false, false));
        assert_ne!(first, Presentation::new(&snapshot(), true, true));
        let mut renamed = snapshot();
        renamed.projects[0].project.display_name = "Renamed".into();
        assert_ne!(first, Presentation::new(&renamed, true, false));
    }
}
