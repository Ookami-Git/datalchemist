import test from 'node:test';
import assert from 'node:assert/strict';

import {
  parseLength,
  parseAspectRatio,
  normalizeSizing,
  clampHeight,
  resolveHeight,
  DEFAULT_HEIGHT_PX,
  DEFAULT_MIN_HEIGHT_PX,
} from './vueFlowSize.js';

test('parseLength accepts numbers, bare digits and px', () => {
  assert.equal(parseLength(320), 320);
  assert.equal(parseLength('420'), 420);
  assert.equal(parseLength('420px'), 420);
  assert.equal(parseLength(' 420PX '), 420);
  assert.equal(parseLength(0), 0);
});

test('parseLength refuses units that cannot be clamped', () => {
  // %, vh et em ne peuvent pas être bornés par min/max côté JS : on retombe sur
  // aspectRatio ou fill, cf. la doc du module.
  for (const value of ['50%', '50vh', '2em', 'auto', '', null, undefined, {}, -10, NaN]) {
    assert.equal(parseLength(value), null, `${JSON.stringify(value)} should be refused`);
  }
});

test('parseAspectRatio accepts ratios and plain numbers', () => {
  assert.equal(parseAspectRatio('16/9'), 16 / 9);
  assert.equal(parseAspectRatio('16:9'), 16 / 9);
  assert.equal(parseAspectRatio(' 4 / 3 '), 4 / 3);
  assert.equal(parseAspectRatio(1.5), 1.5);
  assert.equal(parseAspectRatio('1.5'), 1.5);
});

test('parseAspectRatio rejects degenerate ratios', () => {
  for (const value of ['16/0', '0/9', '0', -2, 'wide', '', null, {}] ) {
    assert.equal(parseAspectRatio(value), null, `${JSON.stringify(value)} should be rejected`);
  }
});

test('normalizeSizing defaults to a fixed height with no lower bound', () => {
  const sizing = normalizeSizing({});
  assert.equal(sizing.mode, 'fixed');
  assert.equal(sizing.fixedHeight, DEFAULT_HEIGHT_PX);
  assert.equal(sizing.minHeight, null);
  assert.equal(sizing.maxHeight, null);
});

test('normalizeSizing detects the fill mode, case-insensitively', () => {
  for (const value of ['fill', 'FILL', ' Fill ']) {
    assert.equal(normalizeSizing({ height: value }).mode, 'fill');
  }
});

test('normalizeSizing detects the aspect mode', () => {
  const sizing = normalizeSizing({ aspectRatio: '16/9' });
  assert.equal(sizing.mode, 'aspect');
  assert.equal(sizing.aspectRatio, 16 / 9);
});

test('fill wins over aspectRatio when both are given', () => {
  const sizing = normalizeSizing({ height: 'fill', aspectRatio: '16/9' });
  assert.equal(sizing.mode, 'fill');
  // Le ratio reste parsé : il redevient utile si le mode est dégradé.
  assert.equal(sizing.aspectRatio, 16 / 9);
});

test('computed modes get a safety floor, fixed mode does not', () => {
  assert.equal(normalizeSizing({ height: 'fill' }).minHeight, DEFAULT_MIN_HEIGHT_PX);
  assert.equal(normalizeSizing({ aspectRatio: 2 }).minHeight, DEFAULT_MIN_HEIGHT_PX);
  assert.equal(normalizeSizing({ height: 300 }).minHeight, null);
  // Une borne explicite reste prioritaire sur le garde-fou.
  assert.equal(normalizeSizing({ height: 'fill', minHeight: '60px' }).minHeight, 60);
});

test('height stays usable as a fallback in computed modes', () => {
  const sizing = normalizeSizing({ aspectRatio: '16/9', height: '250px' });
  assert.equal(sizing.fixedHeight, 250);
  // Avant toute mesure (largeur inconnue), c'est ce repli qui s'applique.
  assert.equal(resolveHeight(sizing, { width: 0 }), 250);
});

test('clampHeight applies max then min so an explicit floor wins', () => {
  assert.equal(clampHeight(500, { maxHeight: 300 }), 300);
  assert.equal(clampHeight(50, { minHeight: 120 }), 120);
  assert.equal(clampHeight(500, { minHeight: 400, maxHeight: 300 }), 400);
  assert.equal(clampHeight(200, {}), 200);
});

test('resolveHeight derives the height from the width in aspect mode', () => {
  const sizing = normalizeSizing({ aspectRatio: '16/9' });
  assert.equal(resolveHeight(sizing, { width: 1600 }), 900);
  assert.equal(resolveHeight(sizing, { width: 320 }), 180);
});

test('resolveHeight uses the measured space in fill mode', () => {
  const sizing = normalizeSizing({ height: 'fill' });
  assert.equal(resolveHeight(sizing, { available: 640 }), 640);
});

test('resolveHeight falls back when a measure is missing or degenerate', () => {
  const fill = normalizeSizing({ height: 'fill' });
  // Colonne repliée ou carte pas encore étirée : on retombe sur le défaut,
  // jamais sur zéro, sinon Vue Flow ne rend rien du tout.
  assert.equal(resolveHeight(fill, { available: null }), DEFAULT_HEIGHT_PX);
  assert.equal(resolveHeight(fill, { available: 0 }), DEFAULT_HEIGHT_PX);
  assert.equal(resolveHeight(fill, { available: -30 }), DEFAULT_HEIGHT_PX);

  const aspect = normalizeSizing({ aspectRatio: 2 });
  assert.equal(resolveHeight(aspect, { width: 0 }), DEFAULT_HEIGHT_PX);
});

test('resolveHeight respects the bounds in every mode', () => {
  const aspect = normalizeSizing({ aspectRatio: '1/2', maxHeight: '500px' });
  assert.equal(resolveHeight(aspect, { width: 600 }), 500);

  const fill = normalizeSizing({ height: 'fill', minHeight: '300px' });
  assert.equal(resolveHeight(fill, { available: 100 }), 300);
});
