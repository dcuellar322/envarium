#[cfg(not(debug_assertions))]
use tauri::plugin::TauriPlugin;
use tauri::{AppHandle, Runtime};
use tauri_plugin_notification::NotificationExt;

#[cfg(test)]
fn validate_endpoint(value: &str) -> Result<url::Url, String> {
    let endpoint = url::Url::parse(value)
        .map_err(|_| "the configured update endpoint is invalid".to_owned())?;
    if endpoint.scheme() != "https"
        || endpoint.host_str().is_none()
        || !endpoint.username().is_empty()
        || endpoint.password().is_some()
        || endpoint.fragment().is_some()
    {
        return Err(
            "the configured update endpoint must be an HTTPS URL without credentials or a fragment"
                .to_owned(),
        );
    }
    Ok(endpoint)
}

#[cfg(not(debug_assertions))]
pub fn plugin<R: Runtime>() -> TauriPlugin<R, tauri_plugin_updater::Config> {
    tauri_plugin_updater::Builder::new().build()
}

pub async fn check_and_install<R: Runtime>(app: AppHandle<R>) {
    #[cfg(debug_assertions)]
    {
        let _ = app
            .notification()
            .builder()
            .title("Envarium updates")
            .body("Update checks are enabled only in signed release builds.")
            .show();
    }

    #[cfg(not(debug_assertions))]
    {
        use tauri_plugin_updater::UpdaterExt;

        let result = async {
            let updater = app.updater().map_err(|error| error.to_string())?;
            let Some(update) = updater.check().await.map_err(|error| error.to_string())? else {
                return Ok::<_, String>(false);
            };
            app.notification()
                .builder()
                .title("Envarium update available")
                .body(format!("Downloading signed version {}…", update.version))
                .show()
                .map_err(|error| error.to_string())?;
            update
                .download_and_install(|_, _| {}, || {})
                .await
                .map_err(|error| error.to_string())?;
            Ok(true)
        }
        .await;

        match result {
            Ok(true) => app.restart(),
            Ok(false) => {
                let _ = app
                    .notification()
                    .builder()
                    .title("Envarium is up to date")
                    .body("No newer signed release is available.")
                    .show();
            }
            Err(error) => {
                let _ = app
                    .notification()
                    .builder()
                    .title("Envarium update failed")
                    .body(error)
                    .show();
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn accepts_https_update_endpoint() {
        let endpoint = validate_endpoint(
            "https://updates.example.com/stable/{{target}}/{{arch}}/{{current_version}}",
        )
        .unwrap();
        assert_eq!(endpoint.scheme(), "https");
    }

    #[test]
    fn release_configuration_has_canonical_feed_and_public_trust() {
        let config: serde_json::Value =
            serde_json::from_str(include_str!("../tauri.release.conf.json")).unwrap();
        let updater = &config["plugins"]["updater"];
        let endpoints = updater["endpoints"].as_array().unwrap();
        assert_eq!(endpoints.len(), 1);
        let endpoint = validate_endpoint(endpoints[0].as_str().unwrap()).unwrap();
        assert_eq!(
            endpoint.as_str(),
            "https://envarium.davidcuellar.tech/updates/stable.json"
        );
        assert!(!updater["pubkey"].as_str().unwrap().is_empty());
        assert_eq!(updater["requireSignedVersion"], true);
        assert_eq!(config["bundle"]["createUpdaterArtifacts"], true);
    }

    #[test]
    fn rejects_unsafe_update_endpoints() {
        for endpoint in [
            "http://updates.example.com/latest.json",
            "https://user:secret@updates.example.com/latest.json",
            "https://updates.example.com/latest.json#unsigned",
            "file:///tmp/latest.json",
        ] {
            assert!(validate_endpoint(endpoint).is_err(), "accepted {endpoint}");
        }
    }
}
