import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { resolve, basename } from 'node:path';
const root = resolve('artifacts/releases');
const sha = (raw) => createHash('sha256').update(raw).digest('hex');
// Immutable published reference: v2.0.0-alpha2/checksums.txt.
const publishedLegacy = new Map([
  ['1.0.5', '80bfee035a035653408aeb746c2d7817b55dcffeb8b13ec8eb3a1a0e2f4845f1'],
]);
test('rebuilt compatibility artifacts never replace a published immutable version', () => {
  const raw = readFileSync(resolve(root, 'docker-module-v1.json'));
  const legacy = JSON.parse(raw);
  if (publishedLegacy.has(legacy.version))
    assert.equal(
      sha(raw),
      publishedLegacy.get(legacy.version),
      'Legacy bytes changed: issue a new compatibility version',
    );
  assert.equal(legacy.schemaVersion, 1);
  for (const binary of Object.values(legacy.binaries))
    assert.equal(sha(Buffer.from(binary.content, 'base64')), binary.sha256);
});
test('release checksums describe the exact emitted artifacts', () => {
  for (const line of readFileSync(resolve(root, 'checksums.txt'), 'utf8').trim().split(/\r?\n/)) {
    const match = /^([a-f0-9]{64})  (.+)$/.exec(line);
    assert.ok(match);
    assert.equal(basename(match[2]), match[2]);
    assert.equal(sha(readFileSync(resolve(root, match[2]))), match[1]);
  }
  const p = JSON.parse(readFileSync(resolve(root, 'docker-module.json'), 'utf8'));
  if (p.schemaVersion === 2 && Number(p.version.split('.')[0]) >= 3) {
    assert.equal(p.components[0].runtime, 'wasip1');
    assert.deepEqual(Object.keys(p.components[0].binaries), ['wasip1-wasm']);
    const b = p.components[0].binaries['wasip1-wasm'];
    const data = Buffer.from(b.content, 'base64');
    assert.deepEqual([...data.subarray(0, 8)], [0, 97, 115, 109, 1, 0, 0, 0]);
    assert.equal(sha(data), b.sha256);
  }
});
