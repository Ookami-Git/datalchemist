# Datalchemist

## Projet

### Backend GO

- Rédige et maintiens à jours les fichiers \_test.go.

## Préférences

- Réponds en français.
- Sois concis ; pas de résumé de ce que tu viens de faire.
- Avant toute modification non triviale, propose un plan court.
- N'invente rien : si tu ne sais pas, dis-le. Vérifie dans le code ou la doc avant d'affirmer, et cite ta source.
- Si ma demande est ambiguë, pose-moi des questions avant d'agir.

## graphify

Ce projet possède un graphe de connaissance dans `graphify-out/`, avec des nœuds centraux, une structure en communautés et des relations inter-fichiers.

Quand l'utilisateur saisit `/graphify`, invoquer le skill `graphify` avant toute autre action.

Règles :

- Pour les questions sur le code, exécuter d'abord `graphify query "<question>"` si `graphify-out/graph.json` existe. Utiliser `graphify path "<A>" "<B>"` pour les relations et `graphify explain "<concept>"` pour un concept ciblé.
- Des fichiers modifiés dans `graphify-out/` sont normaux après les hooks ou mises à jour incrémentales ; ce n'est pas une raison pour ne pas utiliser Graphify. Ne l'ignorer que si la demande concerne un graphe obsolète ou incorrect, ou si l'utilisateur le demande explicitement.
- Si `graphify-out/wiki/index.md` existe, l'utiliser pour une navigation générale plutôt que de parcourir directement les sources.
- Ne lire `graphify-out/GRAPH_REPORT.md` que pour une revue d'architecture large ou si `query`, `path` et `explain` ne fournissent pas assez de contexte.
- Après une modification du code, exécuter `graphify update .` pour garder le graphe à jour (analyse AST uniquement, sans coût d'API).
