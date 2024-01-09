use std::process::{Command};

fn main() {
    set_git_hash();
    set_git_tag();
    set_commit_date();
}

fn set_git_tag() {
    let output = Command::new("git").args(&["describe", "--tags", "--abbrev=0"]).output().unwrap();
    let git_tag = String::from_utf8(output.stdout).unwrap();
    let status_code = output.status.code().unwrap();

    if status_code.eq(&0) {
        println!("cargo:rustc-env=GIT_TAG={}", git_tag);
    }
}

fn set_git_hash() {
    let output = Command::new("git").args(&["rev-parse", "--short", "HEAD"]).output().unwrap();
    let git_hash = String::from_utf8(output.stdout).unwrap();

    println!("cargo:rustc-env=GIT_HASH={}", git_hash);
}

fn set_commit_date() {
    let output = Command::new("git").args(&["show", "-s", "--format=%ci"]).output().unwrap();
    let commit_date = String::from_utf8(output.stdout).unwrap();

    println!("cargo:rustc-env=COMMIT_DATE={}", commit_date);
}