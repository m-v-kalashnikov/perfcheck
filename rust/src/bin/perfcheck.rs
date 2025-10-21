use std::env;
use std::path::PathBuf;
use std::process;

use perfcheck_lint::{lint_path, registry};

fn main() {
    let args: Vec<String> = env::args().collect();
    let target = args
        .get(1)
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."));

    let registry = registry();
    let result = lint_path(&target, registry).unwrap_or_else(|err| {
        eprintln!("error: {}", err);
        process::exit(2);
    });

    if result.is_empty() {
        return;
    }

    for diag in &result {
        println!(
            "{}:{}:{} [{}] {}",
            diag.path.display(),
            diag.line,
            diag.column,
            diag.rule_id,
            diag.message
        );
    }

    process::exit(1);
}
