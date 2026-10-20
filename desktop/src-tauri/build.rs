fn main() {
    println!("cargo:rerun-if-env-changed=ENVARIUM_VERSION");
    let version = std::env::var("ENVARIUM_VERSION")
        .unwrap_or_else(|_| std::env::var("CARGO_PKG_VERSION").expect("Cargo package version"));
    println!("cargo:rustc-env=ENVARIUM_DESKTOP_VERSION={version}");
    tauri_build::build()
}
