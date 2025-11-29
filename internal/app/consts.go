package app

import "time"

const SERVER_MUSIC_JACKET_BASE_URL = "https://maimaidx.jp/maimai-mobile/img/Music/"
const DBFILE = "scores.db"
const MAX_SIMUTANOUS_DOWNLOADS = 16

const MAIMAI_SONGS_JSON_LINK = "https://maimai.sega.jp/data/maimai_songs.json"
const REIWA_JSON_LINK = "https://reiwa.f5.si/maimai_record.json"
const DIVING_FISH_API = "https://www.diving-fish.com/api/maimaidxprober/"
const DIVING_FISH_ETAGS_JSON = "resources/diving_fish_etags.json"

const HTTP_REQUEST_TIMEOUT = 10 * time.Second
