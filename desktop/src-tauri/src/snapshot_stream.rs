use std::time::Duration;

use tauri::{AppHandle, Runtime, async_runtime::Receiver};
use tauri_plugin_shell::{
    ShellExt,
    process::{CommandChild, CommandEvent},
};

use crate::{
    model::{DesktopSnapshot, Envelope},
    sidecar::{SidecarError, validate_envelope},
};

const OUTPUT_LIMIT: usize = 1024 * 1024;

pub(crate) struct SnapshotStream {
    child: Option<CommandChild>,
    events: Receiver<CommandEvent>,
}

// The bridge can read only tray snapshots. Mutations keep their existing
// independent one-shot sidecars and never receive this resource policy.
impl SnapshotStream {
    fn open<R: Runtime>(app: &AppHandle<R>) -> Result<Self, SidecarError> {
        let command = app
            .shell()
            .sidecar("envarium")
            .map_err(|_| SidecarError("bundled snapshot bridge unavailable".into()))?
            .args([
                "--json",
                "--compact",
                "--non-interactive",
                "--no-color",
                "desktop",
                "snapshot",
                "--stream",
            ])
            .env("PATH", crate::environment::sidecar_path())
            .env("GOMAXPROCS", "1")
            .env("GOGC", "100")
            .env("GOMEMLIMIT", "64MiB")
            .set_raw_out(true);
        let (events, child) = command
            .spawn()
            .map_err(|_| SidecarError("snapshot bridge launch failed".into()))?;
        Ok(Self {
            child: Some(child),
            events,
        })
    }

    async fn read(&mut self) -> Result<DesktopSnapshot, SidecarError> {
        self.child
            .as_mut()
            .ok_or_else(|| SidecarError("snapshot bridge closed".into()))?
            .write(b"snapshot\n")
            .map_err(|_| SidecarError("snapshot bridge request failed".into()))?;
        let mut frame = Frame(Vec::new());
        while let Some(event) = self.events.recv().await {
            match event {
                CommandEvent::Stdout(bytes) => {
                    if let Some(value) = frame.push(&bytes)? {
                        return Ok(value);
                    }
                }
                _ => return Err(SidecarError("snapshot bridge disconnected".into())),
            }
        }
        Err(SidecarError("snapshot bridge disconnected".into()))
    }
}

impl Drop for SnapshotStream {
    fn drop(&mut self) {
        if let Some(child) = self.child.take() {
            let _ = child.kill();
        }
    }
}

pub(crate) async fn snapshot<R: Runtime>(
    app: &AppHandle<R>,
    stream: &mut Option<SnapshotStream>,
) -> Result<DesktopSnapshot, SidecarError> {
    if stream.is_none() {
        *stream = Some(SnapshotStream::open(app)?);
    }
    let result = tokio::time::timeout(
        Duration::from_secs(20),
        stream.as_mut().expect("opened bridge").read(),
    )
    .await
    .unwrap_or_else(|_| Err(SidecarError("snapshot bridge timed out".into())));
    if result.is_err() {
        *stream = None;
    }
    result
}

struct Frame(Vec<u8>);
impl Frame {
    fn push(&mut self, bytes: &[u8]) -> Result<Option<DesktopSnapshot>, SidecarError> {
        if self.0.len() + bytes.len() > OUTPUT_LIMIT {
            return Err(SidecarError(
                "snapshot bridge response exceeded the safety limit".into(),
            ));
        }
        self.0.extend_from_slice(bytes);
        let Some(newline) = self.0.iter().position(|byte| *byte == b'\n') else {
            return Ok(None);
        };
        if newline != self.0.len() - 1 {
            return Err(SidecarError(
                "snapshot bridge returned unsolicited output".into(),
            ));
        }
        let envelope: Envelope<DesktopSnapshot> = serde_json::from_slice(&self.0)
            .map_err(|_| SidecarError("snapshot bridge returned an invalid response".into()))?;
        validate_envelope(&envelope, "desktop.snapshot")?;
        Ok(Some(envelope.data))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn response() -> Vec<u8> {
        br#"{"schemaVersion":"envarium.cli/v1","command":"desktop.snapshot","data":{"system":{"apiVersion":"v1","status":"ready","version":"1.0.0","databaseSchemaVersion":19},"projects":[],"workspaces":[],"operations":[],"diagnosticNotifications":[],"warnings":[],"portConflictCount":0}}
"#.to_vec()
    }

    #[test]
    fn fragmented_snapshot_is_validated_once_complete() {
        let bytes = response();
        let mut frame = Frame(Vec::new());
        assert!(frame.push(&bytes[..17]).unwrap().is_none());
        assert_eq!(
            frame.push(&bytes[17..]).unwrap().unwrap().system.status,
            "ready"
        );
    }

    #[test]
    fn invalid_unbounded_or_unsolicited_output_is_rejected() {
        assert!(Frame(Vec::new()).push(b"invalid\n").is_err());
        assert!(
            Frame(Vec::new())
                .push(&vec![b'x'; OUTPUT_LIMIT + 1])
                .is_err()
        );
        let mut doubled = response();
        doubled.extend(response());
        assert!(Frame(Vec::new()).push(&doubled).is_err());
        let wrong = String::from_utf8(response())
            .unwrap()
            .replace("desktop.snapshot", "runtime.operation");
        assert!(Frame(Vec::new()).push(wrong.as_bytes()).is_err());
    }
}
