use std::{env, ffi::OsString, path::Path};

// Finder and launch-at-login do not inherit an interactive shell's PATH.
// Supply stable tool directories to the sidecar without running shell profiles.
pub fn sidecar_path() -> OsString {
    let current = env::var_os("PATH").unwrap_or_default();
    if cfg!(target_os = "macos") {
        let home = env::var_os("HOME");
        macos_tool_path(&current, home.as_deref().map(Path::new))
    } else {
        current
    }
}

fn macos_tool_path(current: &std::ffi::OsStr, home: Option<&Path>) -> OsString {
    let mut directories: Vec<_> = env::split_paths(current)
        .filter(|path| !path.as_os_str().is_empty())
        .collect();
    let mut candidates = vec![
        "/usr/local/bin".into(),
        "/opt/homebrew/bin".into(),
        "/usr/local/opt/node/bin".into(),
        "/opt/homebrew/opt/node/bin".into(),
        "/Applications/Docker.app/Contents/Resources/bin".into(),
        "/usr/bin".into(),
        "/bin".into(),
        "/usr/sbin".into(),
        "/sbin".into(),
    ];
    if let Some(home) = home {
        candidates.extend([
            home.join(".local/bin"),
            home.join(".cargo/bin"),
            home.join("Library/pnpm"),
        ]);
    }
    for directory in candidates {
        if !directories.contains(&directory) {
            directories.push(directory);
        }
    }
    env::join_paths(directories).unwrap_or_else(|_| current.to_os_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn finder_path_includes_local_tools_without_replacing_existing_precedence() {
        let current = env::join_paths(["/custom/bin", "/usr/bin", "/usr/local/bin"]).unwrap();
        let result = macos_tool_path(&current, Some(Path::new("/Users/developer")));
        let paths: Vec<_> = env::split_paths(&result).collect();
        assert_eq!(paths[0], PathBuf::from("/custom/bin"));
        assert_eq!(
            paths
                .iter()
                .filter(|path| path.as_path() == Path::new("/usr/local/bin"))
                .count(),
            1
        );
        assert!(paths.contains(&PathBuf::from("/Users/developer/.local/bin")));
        assert!(paths.contains(&PathBuf::from("/opt/homebrew/bin")));
        assert!(paths.contains(&PathBuf::from(
            "/Applications/Docker.app/Contents/Resources/bin"
        )));
    }

    #[test]
    fn missing_home_does_not_add_relative_user_directories() {
        let result = macos_tool_path(std::ffi::OsStr::new("/usr/bin"), None);
        assert!(env::split_paths(&result).all(|path| path.is_absolute()));
    }

    #[test]
    fn empty_path_retains_system_tools_without_searching_the_working_directory() {
        let result = macos_tool_path(std::ffi::OsStr::new(""), None);
        let paths: Vec<_> = env::split_paths(&result).collect();
        assert!(paths.iter().all(|path| path.is_absolute()));
        assert!(paths.contains(&PathBuf::from("/usr/bin")));
        assert!(paths.contains(&PathBuf::from("/bin")));
    }
}
