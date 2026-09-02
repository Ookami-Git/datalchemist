import assert from 'node:assert/strict';
import test from 'node:test';
import axios from 'axios';

import { DATA_TIMEOUT, DEFAULT_TIMEOUT, installHttpResilience } from './http.js';

// makeClient monte un client isolé dont l'adaptateur est piloté par le test :
// aucune requête ne sort, et chaque appel est enregistré.
function makeClient(respond) {
  const calls = [];
  const client = axios.create({
    adapter: async (config) => {
      calls.push(config);
      return respond(config, calls.length);
    },
  });
  installHttpResilience(client);
  return { client, calls };
}

function ok(config) {
  return { status: 200, statusText: 'OK', data: 'ok', headers: {}, config };
}

// transportFailure reproduit une réponse tronquée : le navigateur ne rend
// aucune réponse HTTP, seulement une erreur de transport.
function transportFailure(config) {
  const error = new Error('Network Error');
  error.config = config;
  error.request = {};
  return Promise.reject(error);
}

function httpFailure(config, status) {
  const error = new Error(`Request failed with status code ${status}`);
  error.config = config;
  error.request = {};
  error.response = { status, statusText: '', data: '', headers: {}, config };
  return Promise.reject(error);
}

test('une erreur de transport sur un GET est rejouée jusqu à réussir', async () => {
  const { client, calls } = makeClient((config, attempt) =>
    attempt < 3 ? transportFailure(config) : ok(config));

  const response = await client.get('/api/views');

  assert.equal(response.data, 'ok');
  assert.equal(calls.length, 3);
});

test('les tentatives sont bornées et la dernière erreur remonte', async () => {
  const { client, calls } = makeClient(transportFailure);

  await assert.rejects(client.get('/api/views'), /Network Error/);
  assert.equal(calls.length, 3);
});

// Rejouer une écriture créerait un doublon ou réappliquerait un import.
test('une écriture n est jamais rejouée', async () => {
  for (const method of ['post', 'put', 'delete']) {
    const { client, calls } = makeClient(transportFailure);
    await assert.rejects(client[method]('/api/item', {}));
    assert.equal(calls.length, 1, `${method} rejoué`);
  }
});

// Forme majoritaire de l'échec sur un lien à faible MTU : le serveur a répondu
// 200, puis le transfert s'est arrêté avant la fin annoncée. axios rend alors
// une erreur qui porte quand même la réponse.
test('un transfert interrompu sous une réponse 200 est rejoué', async () => {
  const { client, calls } = makeClient((config, attempt) => {
    if (attempt >= 2) return ok(config);
    const error = new Error('stream has been aborted');
    error.code = 'ERR_BAD_RESPONSE';
    error.config = config;
    error.request = {};
    error.response = { status: 200, statusText: 'OK', data: '', headers: {}, config };
    return Promise.reject(error);
  });

  const response = await client.get('/assets/index.js');

  assert.equal(response.data, 'ok');
  assert.equal(calls.length, 2);
});

test('seules les erreurs de passerelle sont rejouées', async () => {
  for (const status of [502, 503, 504]) {
    const { client, calls } = makeClient((config, attempt) =>
      attempt < 2 ? httpFailure(config, status) : ok(config));
    await client.get('/api/views');
    assert.equal(calls.length, 2, `${status} non rejoué`);
  }

  for (const status of [400, 401, 403, 404, 500]) {
    const { client, calls } = makeClient((config) => httpFailure(config, status));
    await assert.rejects(client.get('/api/views'));
    assert.equal(calls.length, 1, `${status} rejoué à tort`);
  }
});

// Changer de vue pendant l'attente ne doit pas relancer la requête abandonnée.
test('une annulation interrompt les tentatives', async () => {
  const controller = new AbortController();
  const { client, calls } = makeClient((config) => {
    controller.abort();
    return transportFailure(config);
  });

  await assert.rejects(client.get('/api/views', { signal: controller.signal }));
  assert.equal(calls.length, 1);
});

test('les délais par défaut distinguent les routes de données', async () => {
  const seen = [];
  const { client } = makeClient((config) => {
    seen.push(config.timeout);
    return ok(config);
  });

  await client.get('/api/views');
  await client.get('/api/data/view/1');
  // Un appel qui fixe son propre délai le conserve.
  await client.get('/api/views', { timeout: 1234 });

  assert.deepEqual(seen, [DEFAULT_TIMEOUT, DATA_TIMEOUT, 1234]);
});
