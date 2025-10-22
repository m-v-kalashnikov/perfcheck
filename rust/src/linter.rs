use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::{Path, PathBuf};

use crate::rules::{Rule, RuleRegistry};

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Diagnostic {
    pub rule_id: String,
    pub severity: String,
    pub message: String,
    pub path: PathBuf,
    pub line: usize,
    pub column: usize,
}

pub fn lint_path(path: &Path, registry: &'static RuleRegistry) -> std::io::Result<Vec<Diagnostic>> {
    let mut diagnostics = Vec::new();
    walk(path, registry, &mut diagnostics)?;
    Ok(diagnostics)
}

pub fn lint_source(source: &str, path: &Path, registry: &'static RuleRegistry) -> Vec<Diagnostic> {
    let mut ctx = FileContext::new(path, registry);
    ctx.process(source);
    ctx.into_diagnostics()
}

fn walk(
    path: &Path,
    registry: &'static RuleRegistry,
    acc: &mut Vec<Diagnostic>,
) -> std::io::Result<()> {
    if path.file_name().is_some_and(|name| name == "target") {
        return Ok(());
    }

    let metadata = fs::metadata(path)?;
    if metadata.is_file() {
        if path.extension().is_some_and(|ext| ext == "rs") {
            let source = fs::read_to_string(path)?;
            acc.extend(lint_source(&source, path, registry));
        }
        return Ok(());
    }

    for entry in fs::read_dir(path)? {
        let entry = entry?;
        walk(&entry.path(), registry, acc)?;
    }
    Ok(())
}

#[derive(Default)]
struct Scope {
    vars: HashMap<String, VarInfo>,
}

#[derive(Clone, Debug)]
enum VarInfo {
    String,
    Vector { reserved: bool },
    DynamicDispatch,
}

struct FileContext<'a> {
    path: &'a Path,
    scopes: Vec<Scope>,
    loops: Vec<usize>,
    brace_depth: usize,
    pending_loop: bool,
    diagnostics: Vec<Diagnostic>,
    string_rule: &'static Rule,
    vec_rule: &'static Rule,
    dyn_rule: &'static Rule,
    concurrency_rule: &'static Rule,
    borrow_rule: &'static Rule,
    reported_spawn_lines: HashSet<usize>,
    reported_clone_lines: HashSet<usize>,
}

impl<'a> FileContext<'a> {
    fn new(path: &'a Path, registry: &'static RuleRegistry) -> Self {
        let string_rule = registry
            .rule("perf_avoid_string_concat_loop")
            .expect("perf_avoid_string_concat_loop rule missing");
        let vec_rule = registry
            .rule("perf_vec_reserve_capacity")
            .or_else(|| registry.rule("perf_preallocate_collections"))
            .expect("vector rule missing");
        let dyn_rule = registry
            .rule("perf_avoid_reflection_dynamic")
            .expect("perf_avoid_reflection_dynamic rule missing");
        let concurrency_rule = registry
            .rule("perf_bound_concurrency")
            .expect("perf_bound_concurrency rule missing");
        let borrow_rule = registry
            .rule("perf_borrow_instead_of_clone")
            .expect("perf_borrow_instead_of_clone rule missing");

        Self {
            path,
            scopes: vec![Scope::default()],
            loops: Vec::new(),
            brace_depth: 0,
            pending_loop: false,
            diagnostics: Vec::new(),
            string_rule,
            vec_rule,
            dyn_rule,
            concurrency_rule,
            borrow_rule,
            reported_spawn_lines: HashSet::new(),
            reported_clone_lines: HashSet::new(),
        }
    }

    fn into_diagnostics(self) -> Vec<Diagnostic> {
        self.diagnostics
    }

    fn process(&mut self, source: &str) {
        for (idx, raw_line) in source.lines().enumerate() {
            let line_no = idx + 1;
            let line = raw_line.trim();

            if line.starts_with("for ") || line.starts_with("while ") || line.starts_with("loop") {
                self.pending_loop = true;
            }

            if let Some((name, info)) = parse_variable_declaration(line) {
                if let Some(scope) = self.scopes.last_mut() {
                    scope.vars.insert(name, info);
                }
            }

            if !self.loops.is_empty() {
                self.detect_string_concat(raw_line, line_no);
                self.detect_vector_push(raw_line, line_no);
                self.detect_dynamic_dispatch(raw_line, line_no);
                self.detect_spawn(raw_line, line_no);
                self.detect_clone(raw_line, line_no);
                self.track_vector_reserve(raw_line);
            } else {
                self.track_vector_reserve(raw_line);
            }

            self.process_braces(raw_line);
        }
    }

    fn process_braces(&mut self, raw_line: &str) {
        for ch in raw_line.chars() {
            match ch {
                '{' => {
                    self.brace_depth += 1;
                    self.scopes.push(Scope::default());
                    if self.pending_loop {
                        self.loops.push(self.brace_depth);
                        self.pending_loop = false;
                    }
                }
                '}' => {
                    if self.brace_depth > 0 {
                        if let Some(&depth) = self.loops.last() {
                            if depth == self.brace_depth {
                                self.loops.pop();
                            }
                        }
                        self.scopes.pop();
                        self.brace_depth -= 1;
                    }
                }
                _ => {}
            }
        }
    }

    fn detect_string_concat(&mut self, raw_line: &str, line_no: usize) {
        if let Some(idx) = raw_line.find("+=") {
            let left = raw_line[..idx].trim();
            if self.is_string_var(left) {
                self.push_diag(
                    line_no,
                    idx + 1,
                    left,
                    self.string_rule,
                    "string concatenation in loop",
                );
            }
            return;
        }

        if let Some((lhs, rhs)) = split_assignment(raw_line) {
            let lhs_trim = lhs.trim();
            if !self.is_string_var(lhs_trim) {
                return;
            }
            let rhs_trim = rhs.trim_start();
            if rhs_trim.starts_with(lhs_trim) && rhs_trim.contains('+') {
                self.push_diag(
                    line_no,
                    raw_line.find('=').unwrap_or(0) + 1,
                    lhs_trim,
                    self.string_rule,
                    "string concatenation in loop",
                );
            }
        }
    }

    fn detect_vector_push(&mut self, raw_line: &str, line_no: usize) {
        if let Some(idx) = raw_line.find(".push(") {
            let target = raw_line[..idx].trim_end();
            if let Some(var) = target.rsplit('.').next() {
                let var = var.trim();
                if self.is_vector_without_capacity(var) {
                    self.push_diag(
                        line_no,
                        idx + 1,
                        var,
                        self.vec_rule,
                        "vector push without reserved capacity inside loop",
                    );
                }
            }
        }
    }

    fn track_vector_reserve(&mut self, raw_line: &str) {
        if let Some(idx) = raw_line.find(".reserve(") {
            let target = raw_line[..idx].trim_end();
            if let Some(var) = target.rsplit('.').next() {
                self.mark_vector_reserved(var.trim());
            }
        }
    }

    fn detect_dynamic_dispatch(&mut self, raw_line: &str, line_no: usize) {
        let trimmed = raw_line.trim();
        if trimmed.is_empty() || trimmed.starts_with("//") {
            return;
        }
        let bytes = raw_line.as_bytes();
        let mut i = 0usize;
        while let Some(pos) = raw_line[i..].find('.') {
            let dot_idx = i + pos;
            if dot_idx == 0 {
                i = dot_idx + 1;
                continue;
            }
            let mut start = dot_idx;
            while start > 0 && is_ident_char(bytes[start - 1]) {
                start -= 1;
            }
            if start == dot_idx {
                i = dot_idx + 1;
                continue;
            }
            if !is_ident_start(bytes[start]) {
                i = dot_idx + 1;
                continue;
            }
            let ident = raw_line[start..dot_idx].trim();
            if ident.is_empty() || !self.is_dynamic_var(ident) {
                i = dot_idx + 1;
                continue;
            }

            let mut j = dot_idx + 1;
            while j < bytes.len() && bytes[j].is_ascii_whitespace() {
                j += 1;
            }
            if j >= bytes.len() || !is_ident_start(bytes[j]) {
                i = dot_idx + 1;
                continue;
            }
            while j < bytes.len() && is_ident_char(bytes[j]) {
                j += 1;
            }
            while j < bytes.len() && bytes[j].is_ascii_whitespace() {
                j += 1;
            }
            if j + 1 < bytes.len() && bytes[j] == b':' && bytes[j + 1] == b':' {
                j += 2;
                while j < bytes.len() && bytes[j].is_ascii_whitespace() {
                    j += 1;
                }
                if j < bytes.len() && bytes[j] == b'<' {
                    let mut depth = 1usize;
                    j += 1;
                    while j < bytes.len() && depth > 0 {
                        match bytes[j] {
                            b'<' => depth += 1,
                            b'>' => {
                                depth -= 1;
                            }
                            _ => {}
                        }
                        j += 1;
                    }
                    while j < bytes.len() && bytes[j].is_ascii_whitespace() {
                        j += 1;
                    }
                }
            }
            if j >= bytes.len() || bytes[j] != b'(' {
                i = dot_idx + 1;
                continue;
            }

            self.push_diag(
                line_no,
                dot_idx + 1,
                ident,
                self.dyn_rule,
                "dynamic dispatch inside loop",
            );
            return;
        }
    }

    fn detect_spawn(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_spawn_lines.contains(&line_no) {
            return;
        }
        let trimmed = raw_line.trim();
        if trimmed.is_empty() || trimmed.starts_with("//") {
            return;
        }
        let patterns = [
            "::spawn(",
            ".spawn(",
            "::spawn_blocking(",
            ".spawn_blocking(",
        ];
        if patterns.iter().any(|p| trimmed.contains(p)) {
            self.reported_spawn_lines.insert(line_no);
            self.push_diag(
                line_no,
                raw_line.find("spawn").unwrap_or(1),
                trimmed,
                self.concurrency_rule,
                "spawn inside loop without concurrency bounds",
            );
        }
    }

    fn detect_clone(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_clone_lines.contains(&line_no) {
            return;
        }
        let trimmed = raw_line.trim();
        if trimmed.is_empty() || trimmed.starts_with("//") {
            return;
        }
        let bytes = raw_line.as_bytes();
        let mut index = 0usize;
        while let Some(pos) = raw_line[index..].find(".clone") {
            let start = index + pos;
            let mut cursor = start;
            if cursor == 0 {
                index = start + 6;
                continue;
            }
            while cursor > 0 && is_ident_char(bytes[cursor - 1]) {
                cursor -= 1;
            }
            if cursor >= 2 && bytes[cursor - 1] == b':' && bytes[cursor - 2] == b':' {
                index = start + 6;
                continue;
            }
            if raw_line[start..].starts_with(".clone(") || raw_line[start..].starts_with(".clone()")
            {
                self.reported_clone_lines.insert(line_no);
                self.push_diag(
                    line_no,
                    start + 1,
                    raw_line[cursor..start].trim(),
                    self.borrow_rule,
                    "clone inside loop; prefer borrowing",
                );
                return;
            }
            index = start + 6;
        }
    }

    fn is_string_var(&self, name: &str) -> bool {
        self.lookup_var(name)
            .is_some_and(|info| matches!(info, VarInfo::String))
    }

    fn is_vector_without_capacity(&self, name: &str) -> bool {
        self.lookup_var(name)
            .is_some_and(|info| matches!(info, VarInfo::Vector { reserved: false }))
    }

    fn is_dynamic_var(&self, name: &str) -> bool {
        self.lookup_var(name)
            .is_some_and(|info| matches!(info, VarInfo::DynamicDispatch))
    }

    fn mark_vector_reserved(&mut self, name: &str) {
        for scope in self.scopes.iter_mut().rev() {
            if let Some(VarInfo::Vector { reserved }) = scope.vars.get_mut(name) {
                *reserved = true;
                return;
            }
        }
    }

    fn lookup_var(&self, name: &str) -> Option<&VarInfo> {
        for scope in self.scopes.iter().rev() {
            if let Some(info) = scope.vars.get(name) {
                return Some(info);
            }
        }
        None
    }

    fn push_diag(&mut self, line: usize, column: usize, subject: &str, rule: &Rule, detail: &str) {
        self.diagnostics.push(Diagnostic {
            rule_id: rule.id.clone(),
            severity: rule.severity.clone(),
            message: format!("{} (\"{}\")", detail, subject),
            path: self.path.to_path_buf(),
            line,
            column,
        });
    }
}

fn parse_variable_declaration(line: &str) -> Option<(String, VarInfo)> {
    if !line.starts_with("let ") {
        return None;
    }
    let remainder = line.trim_start_matches("let ").trim_end_matches(';').trim();
    let (left, right) = remainder.split_once('=')?;
    let name = extract_identifier(left)?;
    let init = right.trim();

    if init.contains("String::new()")
        || init.contains("String::from(")
        || init.contains("String::with_capacity")
    {
        return Some((name, VarInfo::String));
    }

    if init.contains("Vec::with_capacity") {
        return Some((name, VarInfo::Vector { reserved: true }));
    }

    if init.contains("Vec::new()") || (init.contains("Vec::<") && init.contains("::new()")) {
        return Some((name, VarInfo::Vector { reserved: false }));
    }

    if contains_trait_object(left) || contains_trait_object(init) {
        return Some((name, VarInfo::DynamicDispatch));
    }

    None
}

fn extract_identifier(left: &str) -> Option<String> {
    let mut ident = left.trim();
    if let Some((before_type, _ty)) = ident.split_once(':') {
        ident = before_type.trim();
    }
    if let Some(stripped) = ident.strip_prefix("mut ") {
        ident = stripped.trim();
    }
    if ident.is_empty() {
        None
    } else {
        Some(ident.to_string())
    }
}

fn split_assignment(line: &str) -> Option<(&str, &str)> {
    if let Some(idx) = line.find('=') {
        let lhs = &line[..idx];
        let rhs = &line[idx + 1..];
        if !lhs.contains('+') {
            return Some((lhs, rhs));
        }
    }
    None
}

fn contains_trait_object(text: &str) -> bool {
    let lower = text.to_ascii_lowercase();
    lower.contains("box<dyn")
        || lower.contains("rc<dyn")
        || lower.contains("arc<dyn")
        || lower.contains("&dyn")
        || lower.contains("dyn ")
        || lower.contains("dyn>")
}

fn is_ident_start(ch: u8) -> bool {
    ch == b'_' || ch.is_ascii_alphabetic()
}

fn is_ident_char(ch: u8) -> bool {
    ch == b'_' || ch.is_ascii_alphanumeric()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::registry;
    use std::collections::HashSet;

    const TEST_FILE: &str = "test.rs";

    #[test]
    fn detects_string_concat_in_loop() {
        let code = r#"
fn build(items: &[String]) -> String {
    let mut out = String::new();
    for item in items {
        out += item;
    }
    out
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert_eq!(diags.len(), 1);
        assert_eq!(diags[0].rule_id, "perf_avoid_string_concat_loop");
    }

    #[test]
    fn detects_vector_push_without_capacity() {
        let code = r#"
fn collect(count: usize) -> Vec<i32> {
    let mut data = Vec::new();
    for i in 0..count {
        data.push(i as i32);
    }
    data
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert_eq!(diags.len(), 1);
        assert_eq!(diags[0].rule_id, "perf_vec_reserve_capacity");
    }

    #[test]
    fn ignores_reserved_vectors() {
        let code = r#"
fn collect(count: usize) -> Vec<i32> {
    let mut data = Vec::with_capacity(count);
    for i in 0..count {
        data.push(i as i32);
    }
    data
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.is_empty());
    }

    #[test]
    fn detects_dynamic_dispatch_in_loop() {
        let code = r#"
trait Handler {
    fn handle(&self, value: i32);
}

fn run(items: &[i32], input: &dyn Handler) {
    let handler: &dyn Handler = input;
    for value in items {
        handler.handle(*value);
    }
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert_eq!(diags.len(), 1);
        assert_eq!(diags[0].rule_id, "perf_avoid_reflection_dynamic");
    }

    #[test]
    fn allows_static_dispatch_in_loop() {
        let code = r#"
trait Handler {
    fn handle(&self, value: i32);
}

struct StaticHandler;

impl Handler for StaticHandler {
    fn handle(&self, value: i32) {
        let _ = value;
    }
}

fn run(items: &[i32]) {
    let handler = StaticHandler;
    for value in items {
        handler.handle(*value);
    }
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.is_empty());
    }

    #[test]
    fn detects_thread_spawn_in_loop() {
        let code = r#"
use std::thread;

fn run(items: &[i32]) {
    for item in items {
        thread::spawn(move || println!("{}", item));
    }
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert_eq!(diags.len(), 1);
        assert_eq!(diags[0].rule_id, "perf_bound_concurrency");
    }

    #[test]
    fn detects_clone_in_loop() {
        let code = r#"
fn process(values: &[String]) {
    for value in values {
        let cloned = value.clone();
        let _ = cloned;
    }
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert_eq!(diags.len(), 1);
        assert_eq!(diags[0].rule_id, "perf_borrow_instead_of_clone");
    }

    #[test]
    fn fixtures_trigger_all_rules() {
        let code = include_str!("../fixtures/violations.rs");
        let registry = registry();
        let diags = lint_source(code, Path::new("fixtures/violations.rs"), registry);
        let got: HashSet<String> = diags.into_iter().map(|d| d.rule_id).collect();
        let expected: HashSet<String> = [
            "perf_avoid_string_concat_loop",
            "perf_vec_reserve_capacity",
            "perf_avoid_reflection_dynamic",
            "perf_bound_concurrency",
            "perf_borrow_instead_of_clone",
        ]
        .into_iter()
        .map(String::from)
        .collect();
        assert_eq!(got, expected);
    }
}
