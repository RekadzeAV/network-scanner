# D-Track Evidence PR Snippet

Готовый короткий блок для вставки в описание PR.

```md
## D-Track Evidence

### Smoke / CI
- [x] Windows smoke: `.\scripts\smoke-d-track-topology-export.ps1` PASS
- [ ] Linux/macOS smoke: `./scripts/smoke-d-track-topology-export.sh` (pending)
- [ ] CI job `D-Track Smoke (Topology Export)` (pending)
- [ ] CI URL: pending

### Export Consistency (`json` vs `graphml`)
- [x] Node-set equivalence: PASS
- [x] Edge-set equivalence: PASS
- [x] GraphML metadata keys present: `source_type`, `confidence`, `evidence`

### Graphviz Fallback
- [x] Without `dot`: command does not fail, fallback JSON is generated, diagnostic message is present
- [ ] With `dot`: direct `png/svg` generation (pending)

### External Compatibility
- [ ] yEd import PASS (pending)
- [ ] Gephi import PASS (pending)

### Residual Risk
- Pending: external imports (yEd/Gephi), CI confirmation, and `dot`-available validation.
```

Полный и детальный статус: [D_TRACK_EVIDENCE_CURRENT.md](D_TRACK_EVIDENCE_CURRENT.md).
