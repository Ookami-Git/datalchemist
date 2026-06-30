const GET_VARIABLE_REGEX = /get\.([A-Za-z_$][A-Za-z0-9_$]*)|get\[['"]([^'"\]]+)['"]\]/g;

export function extractGetVariableNames(value) {
  const input = typeof value === 'string' ? value : JSON.stringify(value ?? {});
  const names = new Set();
  let match;

  while ((match = GET_VARIABLE_REGEX.exec(input)) !== null) {
    names.add(match[1] || match[2]);
  }

  return Array.from(names).sort((a, b) => a.localeCompare(b));
}

export function parseGetQuery(input) {
  const raw = `${input || ''}`.trim();
  if (!raw) return {};

  const normalized = raw.startsWith('?') ? raw.slice(1) : raw;
  const searchParams = new URLSearchParams(normalized);
  const parsed = {};

  for (const [key, value] of searchParams.entries()) {
    if (!key) continue;

    if (Object.prototype.hasOwnProperty.call(parsed, key)) {
      const current = parsed[key];
      parsed[key] = Array.isArray(current) ? [...current, value] : [current, value];
    } else {
      parsed[key] = value;
    }
  }

  return parsed;
}

export function formatGetQuery(params = {}) {
  const searchParams = new URLSearchParams();

  Object.entries(params || {}).forEach(([key, value]) => {
    if (!key) return;

    if (Array.isArray(value)) {
      value.forEach((entry) => searchParams.append(key, entry ?? ''));
    } else {
      searchParams.append(key, value ?? '');
    }
  });

  return searchParams.toString();
}

export function valuesForGetVariables(names, savedValues = {}, defaultValues = {}) {
  const values = {};

  names.forEach((name) => {
    values[name] = savedValues?.[name] ?? defaultValues?.[name] ?? '';
  });

  return values;
}

export function mergeGetVariableDefaults(names, params = {}, fallbackValue = '') {
  const merged = { ...(params || {}) };

  names.forEach((name) => {
    if (!Object.prototype.hasOwnProperty.call(merged, name)) {
      merged[name] = fallbackValue;
    }
  });

  return merged;
}

export function effectiveGetQuery(previewQuery = {}, routeQuery = {}) {
  const hasPreviewQuery = Object.keys(previewQuery || {}).length > 0;
  return hasPreviewQuery ? previewQuery : (routeQuery || {});
}
