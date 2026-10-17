use tauri::{Manager, Runtime, Window};

// Native focus is presentation state only. Daemon observation and tray polling
// continue when the desktop is hidden or another application has focus.
pub fn update<R: Runtime>(window: &Window<R>, active: bool) {
    if let Some(webview) = window.app_handle().get_webview_window(window.label()) {
        let _ = webview.eval(script(active));
    }
}

fn script(active: bool) -> String {
    format!(
        "window.__envariumNativeActive={active};window.dispatchEvent(new Event('envarium:activity'));"
    )
}

#[cfg(test)]
mod tests {
    #[test]
    fn activity_script_contains_only_a_boolean_presentation_fact() {
        assert!(super::script(false).contains("__envariumNativeActive=false;"));
        assert!(super::script(true).contains("__envariumNativeActive=true;"));
    }
}
