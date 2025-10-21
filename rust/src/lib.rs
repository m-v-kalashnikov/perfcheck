pub mod linter;
pub mod rules;

pub use linter::{lint_path, lint_source, Diagnostic};
pub use rules::{Rule, RuleRegistry};

/// Provides access to the lazily initialized global rule registry.
pub fn registry() -> &'static RuleRegistry {
    RuleRegistry::global()
}
