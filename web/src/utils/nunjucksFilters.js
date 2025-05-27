import moment from 'moment';

// Ajoute tous tes filtres dans une fonction d'enregistrement
export function registerNunjucksFilters(env) {
  // Filtre: Trouver un élément dans un tableau par chemin
  env.addFilter("find", (arr, path, value) => {
    for (const obj of arr) {
      let currentObj = obj;
      const keys = path.split(".");
      for (const key of keys) {
        if (currentObj && currentObj.hasOwnProperty(key)) {
          currentObj = currentObj[key];
        } else {
          break;
        }
      }
      if (currentObj === value) return obj;
    }
    return null;
  });

  // Filtre: Convertir une chaîne JSON en objet
  env.addFilter("fromjson", str => {
    try { return JSON.parse(str); }
    catch { return null; }
  });

  // Filtre: Formater une date
  env.addFilter("date", (date, outputformat, inputformat) =>
    moment(date, inputformat).format(outputformat)
  );

  // Filtre: Ajouter ou modifier un attribut dans un objet
  env.addFilter('setAttribute', (dictionary, key, value) => {
    const keys = key.split(/(?<!\\)\./).map(part => part.replace(/\\\./g, '.'));
    let currentObj = dictionary;
    for (let i = 0; i < keys.length - 1; i++) {
      const k = keys[i];
      if (!currentObj.hasOwnProperty(k) || typeof currentObj[k] !== "object") {
        currentObj[k] = {};
      }
      currentObj = currentObj[k];
    }
    currentObj[keys[keys.length - 1]] = value;
    return dictionary;
  });

  // Filtre: Diviser une chaîne en tableau
  env.addFilter("split", (str, separator) => {
    if (typeof str !== "string") return [];
    return str.split(separator);
  });
}