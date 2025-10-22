use std::collections::HashMap;
use std::sync::OnceLock;

const DEFAULT_RULES: &str = include_str!(concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../perfcheck-core/config/default_rules.tsv"
));

static REGISTRY: OnceLock<RuleRegistry> = OnceLock::new();

/// Metadata describing a performance-by-default rule.
#[derive(Debug, Eq, PartialEq)]
pub struct Rule {
    pub id: String,
    pub langs: Vec<String>,
    pub description: String,
    pub category: String,
    pub severity: String,
    pub code: u32,
}

/// Immutable rule registry with fast lookups by id or language.
pub struct RuleRegistry {
    all: Vec<Rule>,
    by_id: HashMap<String, usize>,
    by_lang: HashMap<String, Vec<usize>>,
}

impl RuleRegistry {
    /// Returns a lazily-initialized registry based on the embedded rule data.
    pub fn global() -> &'static RuleRegistry {
        REGISTRY.get_or_init(|| {
            RuleRegistry::from_tsv(DEFAULT_RULES).expect("embedded rule registry is invalid")
        })
    }

    fn from_tsv(data: &str) -> Result<Self, String> {
        let mut all = Vec::with_capacity(16);
        for (line_idx, raw_line) in data.lines().enumerate() {
            if line_idx == 0 {
                continue; // header
            }

            let line = raw_line.trim();
            if line.is_empty() || line.starts_with('#') {
                continue;
            }

            let parts: Vec<&str> = line.split('\t').collect();
            if parts.len() != 5 {
                return Err(format!("invalid field count on line {}", line_idx + 1));
            }

            let langs = parts[1]
                .split(',')
                .filter_map(|lang| {
                    let trimmed = lang.trim().to_ascii_lowercase();
                    if trimmed.is_empty() {
                        None
                    } else {
                        Some(trimmed)
                    }
                })
                .collect::<Vec<_>>();

            let mut rule = Rule {
                id: parts[0].trim().to_string(),
                langs,
                description: parts[2].trim().to_string(),
                category: parts[3].trim().to_ascii_lowercase(),
                severity: parts[4].trim().to_ascii_lowercase(),
                code: 0,
            };

            if rule.id.is_empty() {
                return Err(format!("missing rule id on line {}", line_idx + 1));
            }

            rule.code = hash(&rule.id);
            all.push(rule);
        }

        if all.is_empty() {
            return Err("no rules defined".to_string());
        }

        all.sort_by(|a, b| a.id.cmp(&b.id));

        let mut by_id = HashMap::with_capacity(all.len());
        let mut by_lang: HashMap<String, Vec<usize>> = HashMap::new();

        for (idx, rule) in all.iter().enumerate() {
            by_id.insert(rule.id.clone(), idx);
            for lang in &rule.langs {
                by_lang.entry(lang.clone()).or_default().push(idx);
            }
        }

        for indexes in by_lang.values_mut() {
            indexes.sort_by(|a, b| all[*a].id.cmp(&all[*b].id));
        }

        Ok(Self {
            all,
            by_id,
            by_lang,
        })
    }

    /// Returns the metadata for the supplied identifier, if present.
    pub fn rule(&self, id: &str) -> Option<&Rule> {
        self.by_id.get(id).map(|idx| &self.all[*idx])
    }

    /// Returns an iterator over rules matching the provided language token.
    pub fn rules_for_lang<'a>(&'a self, lang: &str) -> RuleIter<'a> {
        let key = lang.to_ascii_lowercase();
        let indexes = self.by_lang.get(&key).map(|v| v.as_slice()).unwrap_or(&[]);
        RuleIter {
            rules: &self.all,
            indexes,
            pos: 0,
        }
    }

    /// Returns the full rule list.
    pub fn all(&self) -> &[Rule] {
        &self.all
    }
}

/// Iterator over language-matched rules.
pub struct RuleIter<'a> {
    rules: &'a [Rule],
    indexes: &'a [usize],
    pos: usize,
}

impl<'a> Iterator for RuleIter<'a> {
    type Item = &'a Rule;

    fn next(&mut self) -> Option<Self::Item> {
        if self.pos >= self.indexes.len() {
            return None;
        }
        let rule = &self.rules[self.indexes[self.pos]];
        self.pos += 1;
        Some(rule)
    }
}

fn hash(id: &str) -> u32 {
    const OFFSET: u32 = 0x811C9DC5;
    const PRIME: u32 = 0x0100_0193;
    let mut h = OFFSET;
    for b in id.as_bytes() {
        h ^= *b as u32;
        h = h.wrapping_mul(PRIME);
    }
    h
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn registry_loads() {
        let registry = RuleRegistry::from_tsv(DEFAULT_RULES).expect("parse");
        assert!(!registry.all().is_empty());
        let go_rules: Vec<&Rule> = registry.rules_for_lang("go").collect();
        assert!(!go_rules.is_empty());
        assert!(registry.rule("perf_avoid_string_concat_loop").is_some());
    }

    #[test]
    fn iterators_are_deterministic() {
        let registry = RuleRegistry::from_tsv(DEFAULT_RULES).expect("parse");
        let first_run: Vec<String> = registry
            .rules_for_lang("rust")
            .map(|rule| rule.id.clone())
            .collect();
        let second_run: Vec<String> = registry
            .rules_for_lang("rust")
            .map(|rule| rule.id.clone())
            .collect();
        assert_eq!(first_run, second_run);
    }
}
