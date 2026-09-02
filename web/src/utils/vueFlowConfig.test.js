import test from 'node:test';
import assert from 'node:assert/strict';

import { parseFlowConfig, DEFAULT_FILL_TARGET } from './vueFlowConfig.js';
import { DEFAULT_HEIGHT_PX } from './vueFlowSize.js';

test('parseFlowConfig passes nodes and edges through untouched', () => {
  const nodes = [{ id: '1', position: { x: 0, y: 0 }, label: 'Source', sourcePosition: 'right' }];
  const edges = [{ id: 'e1', source: '1', target: '2', animated: true, markerEnd: 'arrowclosed' }];

  const { config } = parseFlowConfig(JSON.stringify({ nodes, edges }));

  // Aucun mapping manuel : toutes les options de nœud/arête restent disponibles.
  assert.deepEqual(config.nodes, nodes);
  assert.deepEqual(config.edges, edges);
});

test('parseFlowConfig routes unknown keys to the VueFlow props', () => {
  const { config } = parseFlowConfig(
    JSON.stringify({
      nodes: [],
      height: '420px',
      background: { variant: 'dots' },
      fitViewOnInit: true,
      snapToGrid: true,
      minZoom: 0.2,
      connectionMode: 'loose',
    })
  );

  assert.equal(config.sizing.fixedHeight, 420);
  assert.deepEqual(config.background, { variant: 'dots' });
  assert.deepEqual(config.flowProps, {
    id: 'dc-flow',
    fitViewOnInit: true,
    snapToGrid: true,
    minZoom: 0.2,
    connectionMode: 'loose',
  });
});

test('parseFlowConfig gives each island a distinct id unless one is provided', () => {
  const generated = parseFlowConfig('{}', { fallbackId: 'dc-flow-2' });
  assert.equal(generated.config.flowProps.id, 'dc-flow-2');

  const explicit = parseFlowConfig('{"id":"pipeline"}', { fallbackId: 'dc-flow-2' });
  assert.equal(explicit.config.flowProps.id, 'pipeline');
});

test('parseFlowConfig defaults sizing and collections', () => {
  const { config } = parseFlowConfig('{}');

  // Vue Flow ne rend rien dans un conteneur de hauteur nulle.
  assert.equal(config.sizing.mode, 'fixed');
  assert.equal(config.sizing.fixedHeight, DEFAULT_HEIGHT_PX);
  assert.equal(config.fillTarget, DEFAULT_FILL_TARGET);
  assert.deepEqual(config.nodes, []);
  assert.deepEqual(config.edges, []);
});

test('parseFlowConfig accepts a numeric height', () => {
  const { config } = parseFlowConfig('{"height":320}');
  assert.equal(config.sizing.fixedHeight, 320);
});

test('parseFlowConfig keeps the sizing keys out of the VueFlow props', () => {
  const { config } = parseFlowConfig(
    JSON.stringify({
      height: 'fill',
      aspectRatio: '16/9',
      minHeight: 200,
      maxHeight: 800,
      fillTarget: '.my-box',
      fitViewOnResize: false,
      snapToGrid: true,
    })
  );

  // Ces clés sont à nous : les laisser passer collerait des attributs parasites
  // sur le DOM de <VueFlow>.
  assert.deepEqual(config.flowProps, { id: 'dc-flow', snapToGrid: true });
  assert.equal(config.sizing.mode, 'fill');
  assert.equal(config.sizing.minHeight, 200);
  assert.equal(config.sizing.maxHeight, 800);
  assert.equal(config.fillTarget, '.my-box');
  assert.equal(config.fitViewOnResize, false);
});

test('fitViewOnResize mirrors fitViewOnInit unless set explicitly', () => {
  assert.equal(parseFlowConfig('{"fitViewOnInit":true}').config.fitViewOnResize, true);
  assert.equal(parseFlowConfig('{}').config.fitViewOnResize, false);
  assert.equal(
    parseFlowConfig('{"fitViewOnInit":true,"fitViewOnResize":false}').config.fitViewOnResize,
    false
  );
});

test('parseFlowConfig ignores a blank fillTarget', () => {
  assert.equal(parseFlowConfig('{"fillTarget":"   "}').config.fillTarget, DEFAULT_FILL_TARGET);
});

test('parseFlowConfig ignores non-array nodes and edges', () => {
  const { config } = parseFlowConfig('{"nodes":"oops","edges":{"a":1}}');
  assert.deepEqual(config.nodes, []);
  assert.deepEqual(config.edges, []);
});

test('parseFlowConfig reports malformed input instead of throwing', () => {
  for (const raw of ['', '   ', '{nope}', '[1,2]', 'null', '"a string"']) {
    const result = parseFlowConfig(raw);
    assert.ok(result.error, `${JSON.stringify(raw)} should be rejected`);
    assert.equal(result.config, undefined);
  }
});

test('parseFlowConfig accepts an already parsed object', () => {
  const { config } = parseFlowConfig({ nodes: [{ id: '1' }] });
  assert.deepEqual(config.nodes, [{ id: '1' }]);
});
