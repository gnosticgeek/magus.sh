//! The boundary between the front end and the `magus` binary.
//!
//! The one rule: **the front end names a command, it never supplies one.**
//!
//! EmuDeck's Electron wrapper exposes `ipcMain.on('bash', (_, cmd) => exec(cmd))`
//! — the renderer hands over a shell string and the main process runs it. For a
//! local app the blast radius is bounded, but it means any front-end bug is
//! arbitrary command execution rather than a rendering glitch. Here the only
//! thing that crosses the boundary is a variant of `Command`, and the mapping
//! from variant to argv lives in this file. There is no code path by which a
//! string from the web view reaches a shell.
//!
//! Note also that nothing here shells out through `sh -c`: `std::process::Command`
//! executes the binary directly with an argument vector, so quoting and word
//! splitting never enter into it.

use std::path::PathBuf;
use std::process::Stdio;

use serde::{Deserialize, Serialize};
use tokio::process::Command as TokioCommand;

/// Every action the front end may request. Adding a capability means adding a
/// variant here, deliberately — which is the point.
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
#[serde(rename_all = "kebab-case")]
pub enum Command {
    /// Read-only: report drift, change nothing.
    Doctor,
    /// Converge the machine to the existing manifest.
    Reconcile,
    /// Write the opinionated manifest, then converge. Used when there is no
    /// manifest yet; the five-question wizard needs a terminal, so the GUI
    /// offers defaults rather than pretending it can ask.
    RunDefaults,
    /// Reverse what magus installed.
    Uninstall,
    /// Versions only.
    Version,
}

impl Command {
    /// The argument vector for this command. Fixed at compile time.
    fn argv(self, dry_run: bool) -> Vec<&'static str> {
        let mut args = match self {
            Command::Doctor => vec!["doctor"],
            Command::Reconcile => vec!["reconcile"],
            Command::RunDefaults => vec!["run", "--defaults"],
            Command::Uninstall => vec!["uninstall"],
            Command::Version => vec!["version"],
        };
        args.push("--json");
        // Doctor is already read-only; passing --dry-run would be noise.
        if dry_run && !matches!(self, Command::Doctor | Command::Version) {
            args.push("--dry-run");
        }
        args
    }

    /// Whether this command modifies the machine. The UI uses it to decide what
    /// needs confirming — knowledge that belongs next to the command list.
    pub fn is_destructive(self) -> bool {
        matches!(self, Command::Reconcile | Command::RunDefaults | Command::Uninstall)
    }
}

#[derive(Debug, thiserror::Error)]
pub enum MagusError {
    #[error("magus is not installed — expected it on PATH or at ~/.local/bin/magus")]
    NotFound,
    #[error("magus could not be run: {0}")]
    Spawn(String),
    #[error("magus produced output that is not a result document: {0}")]
    BadOutput(String),
}

// Tauri needs the error as a string to cross into JavaScript.
impl serde::Serialize for MagusErrorWire {
    fn serialize<S: serde::Serializer>(&self, s: S) -> Result<S::Ok, S::Error> {
        s.serialize_str(&self.0)
    }
}
pub struct MagusErrorWire(pub String);
impl From<MagusError> for MagusErrorWire {
    fn from(e: MagusError) -> Self {
        MagusErrorWire(e.to_string())
    }
}

/// Locate the binary.
///
/// `~/.local/bin` is checked explicitly because it is where the installer puts
/// magus and it is *not* on the PATH of a graphical session — that is a
/// `.bashrc` addition, and nothing launching a desktop entry sources `.bashrc`.
/// Relying on PATH alone would make the GUI work from a terminal and fail from
/// the launcher, which is the most confusing possible failure.
pub fn locate() -> Option<PathBuf> {
    if let Some(home) = dirs::home_dir() {
        let local = home.join(".local/bin/magus");
        if local.is_file() {
            return Some(local);
        }
    }
    which::which("magus").ok()
}

/// Run a named command and return its JSON document, parsed only far enough to
/// confirm it is one. The GUI's TypeScript owns the schema; re-declaring every
/// field here would be a second definition to keep in step.
pub async fn run(command: Command, dry_run: bool) -> Result<serde_json::Value, MagusError> {
    let bin = locate().ok_or(MagusError::NotFound)?;

    let output = TokioCommand::new(&bin)
        .args(command.argv(dry_run))
        // stderr is the human log; the GUI reads stdout. Captured rather than
        // inherited so it cannot scribble over the app's own output.
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .await
        .map_err(|e| MagusError::Spawn(e.to_string()))?;

    // A non-zero exit is normal here: `doctor` exits 1 when it finds drift, and
    // that is a result, not a failure. The document carries the outcome, so it
    // is parsed regardless of the exit status.
    let stdout = String::from_utf8_lossy(&output.stdout);
    serde_json::from_str(&stdout).map_err(|e| {
        let stderr = String::from_utf8_lossy(&output.stderr);
        MagusError::BadOutput(format!("{e}: {}", stderr.trim()))
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn every_command_asks_for_json() {
        for c in [
            Command::Doctor,
            Command::Reconcile,
            Command::RunDefaults,
            Command::Uninstall,
            Command::Version,
        ] {
            assert!(
                c.argv(false).contains(&"--json"),
                "{c:?} must request the machine-readable contract"
            );
        }
    }

    #[test]
    fn dry_run_is_not_passed_where_it_is_meaningless() {
        // doctor is already read-only; version does nothing at all.
        assert!(!Command::Doctor.argv(true).contains(&"--dry-run"));
        assert!(!Command::Version.argv(true).contains(&"--dry-run"));
        assert!(Command::Reconcile.argv(true).contains(&"--dry-run"));
    }

    #[test]
    fn read_only_commands_are_not_marked_destructive() {
        assert!(!Command::Doctor.is_destructive());
        assert!(!Command::Version.is_destructive());
        assert!(Command::Uninstall.is_destructive());
    }
}
