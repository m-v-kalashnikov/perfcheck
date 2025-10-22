#![warn(clippy::all, clippy::pedantic, clippy::nursery, clippy::cargo)]
#![allow(clippy::module_name_repetitions)]

pub mod linter;
pub mod rules;

pub use linter::{lint_path, lint_source, Diagnostic};
pub use rules::{Rule, RuleRegistry};

/// Provides access to the lazily initialized global rule registry.
#[must_use]
pub fn registry() -> &'static RuleRegistry {
    RuleRegistry::global()
}
