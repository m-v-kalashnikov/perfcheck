use std::rc::Rc;
use std::sync::Arc;
use std::thread;

// perf_avoid_string_concat_loop
pub fn string_concat(items: &[String]) -> String {
    let mut out = String::new();
    for item in items {
        out += item; // trigger: perf_avoid_string_concat_loop
    }
    out
}

// perf_vec_reserve_capacity / perf_preallocate_collections
pub fn vec_push(values: &[i32]) -> Vec<i32> {
    let mut data = Vec::new();
    for value in values {
        data.push(*value); // trigger: perf_vec_reserve_capacity
    }
    data
}

// perf_avoid_reflection_dynamic
pub trait Handler {
    fn handle(&self, value: i32);
}

pub fn dyn_dispatch(items: &[i32], handler: &dyn Handler) {
    let handler_ref: &dyn Handler = handler;
    for value in items {
        handler_ref.handle(*value); // trigger: perf_avoid_reflection_dynamic
    }
}

// perf_bound_concurrency
pub fn spawn_all(values: &[i32]) {
    for value in values {
        thread::spawn(move || println!("{}", value)); // trigger: perf_bound_concurrency
    }
}

// perf_borrow_instead_of_clone
pub fn needless_clone(values: &[Rc<String>], extra: Arc<String>) -> usize {
    let mut total = 0;
    for value in values {
        total += (**value).clone().len(); // trigger: perf_borrow_instead_of_clone
    }
    total + extra.clone().len() // trigger: perf_borrow_instead_of_clone
}

// helper to avoid warnings
pub fn touch<T>(value: T) {
    let _ = value;
}
