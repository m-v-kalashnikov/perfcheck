use std::{
    collections::{HashMap, HashSet},
    fs,
    path::{Path, PathBuf},
};

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

/// Traverses the provided path and collects diagnostics for every Rust file
/// encountered.
///
/// # Errors
/// Returns an [`std::io::Error`] when filesystem metadata or file contents
/// cannot be read.
pub fn lint_path(path: &Path, registry: &'static RuleRegistry) -> std::io::Result<Vec<Diagnostic>> {
    let mut diagnostics = Vec::new();
    walk(path, registry, &mut diagnostics)?;
    Ok(diagnostics)
}

#[must_use]
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
    linked_list_rule: &'static Rule,
    enum_rule: &'static Rule,
    arc_rule: &'static Rule,
    mutex_rule: &'static Rule,
    collect_rule: &'static Rule,
    stack_rule: &'static Rule,
    enum_active: bool,
    enum_balance: i32,
    pending_enum: bool,
    reported_spawn_lines: HashSet<usize>,
    reported_clone_lines: HashSet<usize>,
    reported_linked_lines: HashSet<usize>,
    reported_arc_lines: HashSet<usize>,
    reported_mutex_lines: HashSet<usize>,
    reported_collect_lines: HashSet<usize>,
    reported_stack_lines: HashSet<usize>,
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
        let concurrency_rule =
            registry.rule("perf_bound_concurrency").expect("perf_bound_concurrency rule missing");
        let borrow_rule = registry
            .rule("perf_borrow_instead_of_clone")
            .expect("perf_borrow_instead_of_clone rule missing");
        let linked_list_rule =
            registry.rule("perf_avoid_linked_list").expect("perf_avoid_linked_list rule missing");
        let enum_rule =
            registry.rule("perf_large_enum_variant").expect("perf_large_enum_variant rule missing");
        let arc_rule =
            registry.rule("perf_unnecessary_arc").expect("perf_unnecessary_arc rule missing");
        let mutex_rule = registry
            .rule("perf_atomic_for_small_lock")
            .expect("perf_atomic_for_small_lock rule missing");
        let collect_rule =
            registry.rule("perf_needless_collect").expect("perf_needless_collect rule missing");
        let stack_rule =
            registry.rule("perf_prefer_stack_alloc").expect("perf_prefer_stack_alloc rule missing");

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
            linked_list_rule,
            enum_rule,
            arc_rule,
            mutex_rule,
            collect_rule,
            stack_rule,
            enum_active: false,
            enum_balance: 0,
            pending_enum: false,
            reported_spawn_lines: HashSet::new(),
            reported_clone_lines: HashSet::new(),
            reported_linked_lines: HashSet::new(),
            reported_arc_lines: HashSet::new(),
            reported_mutex_lines: HashSet::new(),
            reported_collect_lines: HashSet::new(),
            reported_stack_lines: HashSet::new(),
        }
    }

    fn into_diagnostics(self) -> Vec<Diagnostic> {
        self.diagnostics
    }

    fn process(&mut self, source: &str) {
        for (idx, raw_line) in source.lines().enumerate() {
            let line_no = idx + 1;
            let line = raw_line.trim();

            self.detect_linked_list(raw_line, line_no);
            self.handle_enum_state(raw_line, line_no);
            self.detect_arc_usage(raw_line, line_no);
            self.detect_mutex_primitives(raw_line, line_no);
            self.detect_needless_collect(raw_line, line_no);
            self.detect_stack_alloc(raw_line, line_no);

            if line.starts_with("for ") || line.starts_with("while ") || line.starts_with("loop") {
                self.pending_loop = true;
            }

            if let Some((name, info)) = parse_variable_declaration(line) {
                if let Some(scope) = self.scopes.last_mut() {
                    scope.vars.insert(name, info);
                }
            }

            let in_loop = !self.loops.is_empty();
            if in_loop {
                self.detect_string_concat(raw_line, line_no);
                self.detect_vector_push(raw_line, line_no);
                self.detect_dynamic_dispatch(raw_line, line_no);
                self.detect_spawn(raw_line, line_no);
                self.detect_clone(raw_line, line_no);
            }

            self.track_vector_reserve(raw_line);
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
        let patterns = ["::spawn(", ".spawn(", "::spawn_blocking(", ".spawn_blocking("];
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

    fn detect_linked_list(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_linked_lines.contains(&line_no) {
            return;
        }
        let trimmed = raw_line.trim();
        if trimmed.starts_with("//") {
            return;
        }
        if let Some(idx) = find_token(raw_line, "LinkedList") {
            self.reported_linked_lines.insert(line_no);
            self.push_diag(
                line_no,
                idx + 1,
                "LinkedList",
                self.linked_list_rule,
                "std::collections::LinkedList hurts cache locality",
            );
        }
    }

    fn handle_enum_state(&mut self, raw_line: &str, line_no: usize) {
        let trimmed = raw_line.trim_start();
        if !self.pending_enum && !self.enum_active && starts_enum_decl(trimmed) {
            self.pending_enum = true;
        }

        let (opens, closes) = brace_counts(raw_line);

        if self.pending_enum && opens > 0 {
            self.enum_active = true;
            self.enum_balance = opens - closes;
            self.pending_enum = false;
        } else if self.enum_active {
            self.enum_balance += opens - closes;
        }

        if self.enum_active {
            self.detect_large_enum_variant(raw_line, line_no);
            if self.enum_balance <= 0 {
                self.enum_active = false;
            }
        }
    }

    fn detect_large_enum_variant(&mut self, raw_line: &str, line_no: usize) {
        if raw_line.trim_start().starts_with("//") {
            return;
        }
        if find_large_array_literal(raw_line).is_some() {
            self.push_diag(
                line_no,
                1,
                raw_line.trim(),
                self.enum_rule,
                "enum variant stores a massive payload; box it",
            );
        }
    }

    fn detect_arc_usage(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_arc_lines.contains(&line_no) {
            return;
        }
        if let Some(inner) = extract_generic_argument(raw_line, "Arc") {
            if contains_nonsend_token(&inner) {
                self.reported_arc_lines.insert(line_no);
                self.push_diag(
                    line_no,
                    find_token(raw_line, "Arc").unwrap_or(1),
                    "Arc",
                    self.arc_rule,
                    "Arc wrapping !Send data adds atomic overhead",
                );
            }
        } else if let Some(arg) = extract_paren_argument(raw_line, "Arc::new") {
            if contains_nonsend_token(&arg) {
                self.reported_arc_lines.insert(line_no);
                self.push_diag(
                    line_no,
                    find_token(raw_line, "Arc::new").unwrap_or(1),
                    "Arc::new",
                    self.arc_rule,
                    "Arc wrapping !Send data adds atomic overhead",
                );
            }
        }
    }

    fn detect_mutex_primitives(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_mutex_lines.contains(&line_no) {
            // still allow second heuristic on same line if first matched
        }
        let mut flagged = false;
        if let Some(inner) = extract_generic_argument(raw_line, "Mutex") {
            if is_primitive_token(inner.trim()) {
                flagged = true;
            }
        }
        if !flagged {
            if let Some(arg) = extract_paren_argument(raw_line, "Mutex::new") {
                if is_literal_primitive(arg.trim()) {
                    flagged = true;
                }
            }
        }
        if flagged {
            self.reported_mutex_lines.insert(line_no);
            self.push_diag(
                line_no,
                find_token(raw_line, "Mutex").unwrap_or(1),
                "Mutex",
                self.mutex_rule,
                "Mutex guarding a primitive should be an atomic",
            );
        }
    }

    fn detect_needless_collect(&mut self, raw_line: &str, line_no: usize) {
        if self.reported_collect_lines.contains(&line_no) {
            return;
        }
        let collapsed = normalized_no_ws(raw_line);
        if let Some(suffix) = find_needless_collect_suffix(&collapsed) {
            self.reported_collect_lines.insert(line_no);
            let col = raw_line.find("collect").map_or(1, |idx| idx + 1);
            let detail = format!("collect::<Vec<_>>() followed by {suffix}");
            self.push_diag(line_no, col, "collect", self.collect_rule, &detail);
        }
    }

    fn detect_stack_alloc(&mut self, raw_line: &str, line_no: usize) {
        if let Some(arg) = extract_paren_argument(raw_line, "Box::new") {
            if is_small_box_target(arg.trim()) {
                self.reported_stack_lines.insert(line_no);
                self.push_diag(
                    line_no,
                    find_token(raw_line, "Box::new").unwrap_or(1),
                    "Box::new",
                    self.stack_rule,
                    "Boxing a small Copy-sized value adds needless heap indirection",
                );
                return;
            }
        }
        if raw_line.contains("Rc::new(") || raw_line.contains("Arc::new(") {
            if let Some(arg) = extract_paren_argument(raw_line, "Rc::new")
                .or_else(|| extract_paren_argument(raw_line, "Arc::new"))
            {
                if is_small_box_target(arg.trim()) {
                    self.reported_stack_lines.insert(line_no);
                    let token = if raw_line.contains("Rc::new") { "Rc::new" } else { "Arc::new" };
                    self.push_diag(
                        line_no,
                        find_token(raw_line, token).unwrap_or(1),
                        token,
                        self.stack_rule,
                        "Heap-indirection for a tiny value is unnecessary",
                    );
                }
            }
        }
    }

    fn is_string_var(&self, name: &str) -> bool {
        self.lookup_var(name).is_some_and(|info| matches!(info, VarInfo::String))
    }

    fn is_vector_without_capacity(&self, name: &str) -> bool {
        self.lookup_var(name)
            .is_some_and(|info| matches!(info, VarInfo::Vector { reserved: false }))
    }

    fn is_dynamic_var(&self, name: &str) -> bool {
        self.lookup_var(name).is_some_and(|info| matches!(info, VarInfo::DynamicDispatch))
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
        let detail_sentence = ensure_sentence(&format!("{detail} (\"{subject}\")"));
        let summary = ensure_sentence(&rule.problem_summary);
        let fix = ensure_sentence(&rule.fix_hint);
        self.diagnostics.push(Diagnostic {
            rule_id: rule.id.clone(),
            severity: rule.severity.clone(),
            message: format!("{detail_sentence} Why: {summary} Fix: {fix}"),
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

    if init.contains("String::new()") ||
        init.contains("String::from(") ||
        init.contains("String::with_capacity")
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
    lower.contains("box<dyn") ||
        lower.contains("rc<dyn") ||
        lower.contains("arc<dyn") ||
        lower.contains("&dyn") ||
        lower.contains("dyn ") ||
        lower.contains("dyn>")
}

fn ensure_sentence(text: &str) -> String {
    let trimmed = text.trim();
    if trimmed.is_empty() {
        return String::new();
    }
    if trimmed.ends_with('.') || trimmed.ends_with('!') || trimmed.ends_with('?') {
        trimmed.to_string()
    } else {
        format!("{trimmed}.")
    }
}

const fn is_ident_start(ch: u8) -> bool {
    ch == b'_' || ch.is_ascii_alphabetic()
}

const fn is_ident_char(ch: u8) -> bool {
    ch == b'_' || ch.is_ascii_alphanumeric()
}

fn find_token(line: &str, token: &str) -> Option<usize> {
    for (idx, _) in line.match_indices(token) {
        let before = line[..idx].chars().rev().find(|ch| !ch.is_whitespace());
        let after = line[idx + token.len()..].chars().find(|ch| !ch.is_whitespace());
        let boundary_before = before.is_none_or(|ch| !ch.is_ascii_alphanumeric() && ch != '_');
        let boundary_after = after.is_none_or(|ch| !ch.is_ascii_alphanumeric() && ch != '_');
        if boundary_before && boundary_after {
            return Some(idx + 1);
        }
    }
    None
}

fn extract_generic_argument(line: &str, ident: &str) -> Option<String> {
    let pattern = format!("{ident}<");
    let start_idx = line.find(&pattern)? + pattern.len();
    let mut depth = 1i32;
    let chars: Vec<char> = line.chars().collect();
    let mut pos = start_idx;
    while pos < chars.len() {
        match chars[pos] {
            '<' => depth += 1,
            '>' => {
                depth -= 1;
                if depth == 0 {
                    return Some(chars[start_idx..pos].iter().collect());
                }
            }
            _ => {}
        }
        pos += 1;
    }
    None
}

fn extract_paren_argument(line: &str, ident: &str) -> Option<String> {
    let pattern = format!("{ident}(");
    let start_idx = line.find(&pattern)? + pattern.len();
    let mut depth = 1i32;
    let chars: Vec<char> = line.chars().collect();
    let mut pos = start_idx;
    while pos < chars.len() {
        match chars[pos] {
            '(' => depth += 1,
            ')' => {
                depth -= 1;
                if depth == 0 {
                    return Some(chars[start_idx..pos].iter().collect());
                }
            }
            _ => {}
        }
        pos += 1;
    }
    None
}

fn find_large_array_literal(line: &str) -> Option<usize> {
    let bytes = line.as_bytes();
    let mut idx = 0usize;
    while idx < bytes.len() {
        if bytes[idx] == b'[' {
            if let Some(semi_rel) = line[idx..].find(';') {
                let semi_idx = idx + semi_rel;
                if let Some(end_rel) = line[semi_idx..].find(']') {
                    let end_idx = semi_idx + end_rel;
                    let number = line[semi_idx + 1..end_idx].trim();
                    if let Ok(value) = number.parse::<u32>() {
                        if value >= 128 {
                            return Some(idx + 1);
                        }
                    }
                }
            }
        }
        idx += 1;
    }
    None
}

fn contains_nonsend_token(text: &str) -> bool {
    let lower = text.to_ascii_lowercase();
    if lower.contains("refcell") || lower.contains("unsafecell") || lower.contains("cell<") {
        return true;
    }
    for (idx, _) in lower.match_indices("rc<") {
        if idx == 0 || lower.as_bytes()[idx - 1] != b'a' {
            return true;
        }
    }
    for (idx, _) in lower.match_indices("rc::") {
        if idx == 0 || lower.as_bytes()[idx - 1] != b'a' {
            return true;
        }
    }
    false
}

fn is_primitive_token(text: &str) -> bool {
    let trimmed = text.trim();
    if trimmed.is_empty() {
        return false;
    }
    if trimmed.starts_with("Option<") && trimmed.ends_with('>') {
        let inner = &trimmed[7..trimmed.len() - 1];
        return is_primitive_token(inner);
    }
    matches!(
        trimmed,
        "bool" |
            "char" |
            "i8" |
            "i16" |
            "i32" |
            "i64" |
            "i128" |
            "isize" |
            "u8" |
            "u16" |
            "u32" |
            "u64" |
            "u128" |
            "usize" |
            "f32" |
            "f64"
    )
}

fn is_literal_primitive(arg: &str) -> bool {
    let trimmed = arg.trim();
    if trimmed.eq("true") || trimmed.eq("false") {
        return true;
    }
    if trimmed.starts_with('\'') && trimmed.ends_with('\'') {
        return true;
    }
    let mut digits = String::new();
    for ch in trimmed.chars() {
        if ch.is_ascii_digit() || ch == '_' {
            digits.push(ch);
            continue;
        }
        break;
    }
    !digits.is_empty()
}

fn normalized_no_ws(line: &str) -> String {
    line.chars().filter(|ch| !ch.is_whitespace()).collect()
}

const NEEDLESS_COLLECT_SUFFIXES: [&str; 7] =
    [".len()", ".is_empty()", ".iter()", ".iter_mut()", ".into_iter()", ".first()", ".last()"];

fn find_needless_collect_suffix(collapsed: &str) -> Option<&'static str> {
    let pattern = ".collect::<Vec<";
    let mut idx = 0usize;
    while let Some(found) = collapsed[idx..].find(pattern) {
        let mut pos = idx + found + pattern.len();
        let bytes = collapsed.as_bytes();
        let mut depth = 1i32;
        while pos < bytes.len() && depth > 0 {
            match bytes[pos] {
                b'<' => depth += 1,
                b'>' => depth -= 1,
                _ => {}
            }
            pos += 1;
        }
        if depth != 0 || pos >= bytes.len() {
            break;
        }
        if !collapsed[pos..].starts_with(">()") {
            idx = pos;
            continue;
        }
        let tail = &collapsed[pos + 3..];
        for suffix in NEEDLESS_COLLECT_SUFFIXES {
            if tail.starts_with(suffix) {
                return Some(suffix);
            }
        }
        idx = pos;
    }
    None
}

fn is_small_box_target(arg: &str) -> bool {
    if is_literal_primitive(arg) {
        return true;
    }
    if arg.starts_with('(') && arg.ends_with(')') {
        return count_commas_outside_braces(arg) <= 2;
    }
    if arg.contains('{') && arg.contains('}') {
        let commas = count_commas_outside_braces(arg);
        return commas <= 2 && arg.len() <= 80;
    }
    false
}

fn count_commas_outside_braces(text: &str) -> usize {
    let mut depth = 0i32;
    let mut count = 0usize;
    for ch in text.chars() {
        match ch {
            '{' | '(' | '[' => depth += 1,
            '}' | ')' | ']' => {
                if depth > 0 {
                    depth -= 1;
                }
            }
            ',' if depth <= 1 => count += 1,
            _ => {}
        }
    }
    count
}

fn brace_counts(line: &str) -> (i32, i32) {
    let mut opens = 0i32;
    let mut closes = 0i32;
    for ch in line.chars() {
        if ch == '{' {
            opens = opens.saturating_add(1);
        } else if ch == '}' {
            closes = closes.saturating_add(1);
        }
    }
    (opens, closes)
}

fn starts_enum_decl(line: &str) -> bool {
    let mut text = line.trim_start();
    if let Some(rest) = text.strip_prefix("pub(crate)") {
        text = rest.trim_start();
    } else if let Some(rest) = text.strip_prefix("pub(super)") {
        text = rest.trim_start();
    } else if let Some(rest) = text.strip_prefix("pub(self)") {
        text = rest.trim_start();
    }
    if let Some(rest) = text.strip_prefix("pub") {
        text = rest.trim_start();
    }
    text.starts_with("enum ")
}

#[cfg(test)]
mod tests {
    use std::collections::HashSet;

    use super::*;
    use crate::registry;

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
        assert!(diags[0].message.contains("Why:"));
        assert!(diags[0].message.contains("Fix:"));
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
    fn detects_linked_list_usage() {
        let code = r#"
use std::collections::LinkedList;

fn build(values: &[i32]) -> LinkedList<i32> {
    let mut list = LinkedList::new();
    for value in values {
        list.push_back(*value);
    }
    list
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_avoid_linked_list"));
    }

    #[test]
    fn detects_large_enum_variant() {
        let code = r#"
enum Payload {
    Thin(u8),
    Fat([u8; 2048]),
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_large_enum_variant"));
    }

    #[test]
    fn detects_unnecessary_arc() {
        let code = r#"
use std::{cell::RefCell, sync::Arc};

fn wrap(value: RefCell<String>) -> Arc<RefCell<String>> {
    Arc::new(value)
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_unnecessary_arc"));
    }

    #[test]
    fn detects_mutex_primitives() {
        let code = r#"
use std::sync::Mutex;

fn make_flag() -> Mutex<bool> {
    Mutex::new(true)
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_atomic_for_small_lock"));
    }

    #[test]
    fn detects_needless_collect() {
        let code = r#"
fn count(items: &[i32]) -> usize {
    items.iter().filter(|v| **v > 0).collect::<Vec<_>>().len()
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_needless_collect"));
    }

    #[test]
    fn detects_small_box_allocations() {
        let code = r#"
fn boxed_point(x: i32, y: i32) -> Box<(i32, i32)> {
    Box::new((x, y))
}
"#;
        let registry = registry();
        let diags = lint_source(code, Path::new(TEST_FILE), registry);
        assert!(diags.iter().any(|d| d.rule_id == "perf_prefer_stack_alloc"));
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
            "perf_avoid_linked_list",
            "perf_large_enum_variant",
            "perf_unnecessary_arc",
            "perf_atomic_for_small_lock",
            "perf_needless_collect",
            "perf_prefer_stack_alloc",
        ]
        .into_iter()
        .map(String::from)
        .collect();
        assert_eq!(got, expected);
    }
}
