// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

use criterion::{criterion_group, criterion_main, Criterion};
use soroban_env_host::{
    budget::Budget,
    storage::{AccessType, Footprint, Storage},
    xdr::{
        AccountEntry, AccountId, LedgerEntry, LedgerEntryData, LedgerEntryExt, LedgerKey,
        LedgerKeyAccount, PublicKey, SequenceNumber, String32, StringM, Thresholds, Uint256,
    },
    DiagnosticLevel, Host, LedgerInfo, MeteredOrdMap,
};
use std::rc::Rc;
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
            signers: soroban_env_host::xdr::VecM::default(),
            ext: soroban_env_host::xdr::AccountEntryExt::V0,
        }),
        ext: LedgerEntryExt::V0,
    };

    (key, entry)
}

fn create_host(entry_count: u32) -> Host {
    let budget = Budget::default();
    let entries: Vec<(Rc<LedgerKey>, Rc<LedgerEntry>)> = (0..entry_count)
        .map(|i| {
            let (key, entry) = create_dummy_account(i);
            (Rc::new(key), Rc::new(entry))
        })
        .collect();

    let footprint = Footprint(
        MeteredOrdMap::from_exact_iter(
            entries
                .iter()
                .map(|(key, _)| (Rc::clone(key), AccessType::ReadOnly)),
            &budget,
        )
        .unwrap(),
    );
    let storage_map = MeteredOrdMap::from_exact_iter(
        entries
            .iter()
            .map(|(key, entry)| (Rc::clone(key), Some((Rc::clone(entry), None)))),
        &budget,
    )
    .unwrap();

    let host = Host::with_storage_and_budget(
        Storage::with_enforcing_footprint_and_map(footprint, storage_map),
        budget,
    );
    host.set_ledger_info(LedgerInfo {
        protocol_version: 25,
        sequence_number: 1,
        timestamp: 0,
        network_id: [0; 32],
        base_reserve: 0,
        min_persistent_entry_ttl: 4096,
        min_temp_entry_ttl: 16,
        max_entry_ttl: 6_312_000,
    })
    .unwrap();
    host.set_diagnostic_level(DiagnosticLevel::Debug).unwrap();
    host
}

fn snapshot_overhead_benchmark(c: &mut Criterion) {
    let mut group = c.benchmark_group("snapshot_overhead");

    // Benchmark 1000 host function calls WITHOUT snapshots
    group.bench_function("no_snapshots", |b| {
        b.iter_with_setup(
            || create_host(0),
            |host: Host| {
                for _ in 0..1000 {
                    let _ = host.get_ledger_protocol_version().unwrap();
                }
            },
        );
    });

    // Benchmark 1000 host function calls WITH snapshots (1000 injected entries)
    group.bench_function("with_snapshots", |b| {
        b.iter_with_setup(
            || create_host(1000),
            |host: Host| {
                for _ in 0..1000 {
                    let _ = host.get_ledger_protocol_version().unwrap();
                }
            },
        );
    });

    group.finish();
}

criterion_group!(benches, snapshot_overhead_benchmark);
criterion_main!(benches);
