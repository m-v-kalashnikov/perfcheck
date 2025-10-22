#![warn(clippy::all, clippy::pedantic, clippy::nursery, clippy::cargo)]
#![allow(clippy::missing_errors_doc, clippy::missing_panics_doc, clippy::module_name_repetitions)]

use std::{
    env, io,
    io::ErrorKind,
    path::PathBuf,
    process::{self, Command, ExitStatus},
};

fn main() {
    if let Err(code) = run() {
        process::exit(code);
    }
}

fn run() -> Result<(), i32> {
    let args: Vec<String> = env::args().skip(1).collect();
    let (clippy_args, perfcheck_target) = split_args(args);

    let clippy_status =
        Command::new("cargo").arg("clippy").args(&clippy_args).status().map_err(|err| {
            eprintln!("failed to invoke cargo clippy: {err}");
            2
        })?;

    if !clippy_status.success() {
        return Err(clippy_status.code().unwrap_or(1));
    }

    let target =
        perfcheck_target.or_else(|| manifest_dir(&clippy_args)).unwrap_or_else(|| ".".to_owned());

    let perfcheck_status = run_perfcheck(&target).map_err(|err| {
        eprintln!("failed to run perfcheck: {err}");
        2
    })?;

    if !perfcheck_status.success() {
        return Err(perfcheck_status.code().unwrap_or(1));
    }

    Ok(())
}

fn split_args(args: Vec<String>) -> (Vec<String>, Option<String>) {
    let mut clippy_args = Vec::with_capacity(args.len());
    let mut perfcheck_target = None;

    let mut iter = args.into_iter();
    while let Some(arg) = iter.next() {
        if arg == "--perfcheck-target" {
            if let Some(value) = iter.next() {
                perfcheck_target = Some(value);
            }
            continue;
        }

        if let Some(value) = arg.strip_prefix("--perfcheck-target=") {
            perfcheck_target = Some(value.to_owned());
            continue;
        }

        clippy_args.push(arg);
    }

    (clippy_args, perfcheck_target)
}

fn manifest_dir(args: &[String]) -> Option<String> {
    let mut iter = args.iter();
    while let Some(arg) = iter.next() {
        if arg == "--manifest-path" {
            if let Some(path) = iter.next() {
                return manifest_parent(path);
            }
            break;
        }

        if let Some(path) = arg.strip_prefix("--manifest-path=") {
            return manifest_parent(path);
        }
    }

    None
}

fn manifest_parent(path: &str) -> Option<String> {
    let mut buf = PathBuf::from(path);
    if !buf.pop() {
        return None;
    }

    Some(buf.to_string_lossy().into_owned())
}

fn run_perfcheck(target: &str) -> io::Result<ExitStatus> {
    match Command::new("perfcheck").arg(target).status() {
        Ok(status) => Ok(status),
        Err(err) if err.kind() == ErrorKind::NotFound => Command::new("cargo")
            .arg("run")
            .arg("--quiet")
            .arg("--bin")
            .arg("perfcheck")
            .arg("--")
            .arg(target)
            .status(),
        Err(err) => Err(err),
    }
}
