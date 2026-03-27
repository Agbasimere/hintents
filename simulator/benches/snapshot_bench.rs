// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

use criterion::{criterion_group, criterion_main, BenchmarkId, Criterion};
use simulator::runner::SimHost;
use soroban_env_host::xdr::{
    AccountEntry, AccountId, LedgerEntry, LedgerEntryData, LedgerEntryExt, LedgerKey,
    LedgerKeyAccount, PublicKey, SequenceNumber, Thresholds, Uint256, String32, StringM,
};
use std::str::FromStr;

/// Helper to create a dummy account key and entry for benchmarking
fn create_dummy_account(i: u32) -> (LedgerKey, LedgerEntry) {
    let mut bytes = [0u8; 32];
    let i_bytes = i.to_be_bytes();
    bytes[28..32].copy_from_slice(&i_bytes);
    let account_id = AccountId(PublicKey::PublicKeyTypeEd25519(Uint256(bytes)));

    let key = LedgerKey::Account(LedgerKeyAccount {
        account_id: account_id.clone(),
    });

    let entry = LedgerEntry {
        last_modified_ledger_seq: 1,
        data: LedgerEntryData::Account(AccountEntry {
            account_id,
            balance: 1000,
            seq_num: SequenceNumber(1),
            num_sub_entries: 0,
            inflation_dest: None,
            flags: 0,
            home_domain: String32(StringM::from_str("bench.com").unwrap()),
            thresholds: Thresholds([1, 0, 0, 0]),
            signers: Default::default(),
            ext: soroban_env_host::xdr::AccountEntryExt::V0,
        }),
        ext: LedgerEntryExt::V0,
    };

    (key, entry)
}

fn snapshot_overhead_benchmark(c: &mut Criterion) {
    let mut group = c.benchmark_group("snapshot_overhead");
    
    // Benchmark 1000 host function calls WITHOUT snapshots
    group.bench_function("no_snapshots", |b| {
        b.iter_with_setup(
            || SimHost::new(None, None, None).inner,
            |host| {
                for _ in 0..1000 {
                    // Call get_ledger_version as a representative "core" host function
                    let _ = host.get_ledger_version().unwrap();
                }
            },
        );
    });

    // Benchmark 1000 host function calls WITH snapshots (1000 injected entries)
    group.bench_function("with_snapshots", |b| {
        b.iter_with_setup(
            || {
                let host = SimHost::new(None, None, None).inner;
                // Inject 1000 ledger entries to simulate a snapshot
                for i in 0..1000 {
                    let (key, entry) = create_dummy_account(i);
                    // Use the injection method used in the codebase tests
                    let _ = host.put_ledger_entry(key, entry);
                }
                host
            },
            |host| {
                for _ in 0..1000 {
                    let _ = host.get_ledger_version().unwrap();
                }
            },
        );
    });

    group.finish();
}

criterion_group!(benches, snapshot_overhead_benchmark);
criterion_main!(benches);
