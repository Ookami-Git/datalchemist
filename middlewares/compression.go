package middlewares

import (
	"path"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// alreadyCompressed liste les extensions dont le format porte déjà sa propre
// compression : les repasser au gzip coûte du CPU pour quelques octets, et
// parfois en ajoute.
var alreadyCompressed = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".avif": {},
	".ico": {}, ".woff": {}, ".woff2": {}, ".ttf": {},
	".zip": {}, ".gz": {}, ".br": {}, ".mp4": {}, ".webm": {},
}

// uncompressedPaths liste les routes dont le corps est déjà une archive
// (`handlers.Export` renvoie un zip) : l'extension n'apparaît pas dans l'URL,
// seul le chemin permet de les reconnaître.
var uncompressedPaths = map[string]struct{}{
	"/api/export": {},
}

// Compression compresse les réponses en gzip quand le client l'accepte. Sur un
// lien dégradé (VPN à faible MTU, perte de paquets), le nombre d'octets sur le
// fil détermine directement le taux d'échec des requêtes : le bundle front
// tombe de ~4,5 Mo à ~1,8 Mo, les réponses JSON d'un facteur 10 environ.
//
// À placer avant static.Serve pour que les assets embarqués en bénéficient
// aussi, et non les seules routes /api.
//
// Le seuil minimal reste volontairement à zéro (pas de gzip.WithMinLength) :
// sous ce seuil, le writer de gin-contrib/gzip accumule la réponse dans un
// tampon que Flush() ne vide pas, ce qui gèlerait les flux SSE de
// `/api/data/.../stream` jusqu'à la fin du chargement. Compresser dès le
// premier octet garde le flux progressif — et c'est justement là que passent
// les données de vue, les plus volumineuses. En échange, une réponse de
// quelques octets grossit de l'entête gzip (17 octets deviennent 41), ce qui
// ne change rien : elle tenait dans un paquet, elle y tient toujours.
func Compression() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression, gzip.WithCustomShouldCompressFn(shouldCompress))
}

// shouldCompress remplace entièrement la décision par défaut de la librairie :
// dès qu'on lui fournit cette fonction, ses propres exclusions ne sont plus
// consultées, y compris la négociation Accept-Encoding.
func shouldCompress(c *gin.Context) bool {
	request := c.Request

	if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	// Un protocole négocié après l'upgrade (WebSocket) ne transporte plus de
	// corps HTTP à compresser.
	if strings.Contains(request.Header.Get("Connection"), "Upgrade") {
		return false
	}
	// Une requête Range est servie par http.ServeContent, qui découpe le
	// contenu brut : compresser la tranche renvoyée décalerait les offsets
	// annoncés dans Content-Range.
	if request.Header.Get("Range") != "" {
		return false
	}
	if _, found := alreadyCompressed[strings.ToLower(path.Ext(request.URL.Path))]; found {
		return false
	}
	if _, found := uncompressedPaths[request.URL.Path]; found {
		return false
	}

	return true
}
