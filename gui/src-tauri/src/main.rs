// Prevents a console window appearing alongside the app on Windows.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod magus;

use magus::{Command, MagusErrorWire};

/// Whether a magus binary is reachable. The front end asks first so it can fall
/// back to fixtures rather than showing an empty window.
#[tauri::command]
fn magus_available() -> bool {
    magus::locate().is_some()
}

/// The single entry point from the web view.
///
/// `command` deserialises into the `Command` enum, so anything that is not a
/// known variant is rejected by serde before it reaches this body. A string is
/// never treated as a command line.
#[tauri::command]
async fn magus_run(command: Command, dry_run: bool) -> Result<serde_json::Value, MagusErrorWire> {
    magus::run(command, dry_run).await.map_err(Into::into)
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![magus_available, magus_run])
        .run(tauri::generate_context!())
        .expect("error while running magus");
}
