package tinixml

import (
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
	"strconv"
	"strings"
)

var awsVoiceId = map[string]map[string]string{
	"en-US": {}, "en-GB": {}, "de-DE": {},
	"en-AU": {}, "en-CA": {}, "en-IN": {},
	"sv-SE": {}, "zh-CN": {}, "pl-PL": {},
	"pt-BR": {}, "fr-FR": {}, "ja-JP": {},
	"pt-PT": {}, "it-IT": {}, "ko-KR": {},
	"ru-RU": {}, "nb-NO": {}, "nl-NL": {},
	"fr-CA": {}, "ca-ES": {}, "es-ES": {},
	"es-MX": {}, "fi-FI": {}, "zh-HK": {},
	"zh-TW": {}, "da-DK": {},
}

var validAwsPollyVoiceId = "Aditi | Amy | Astrid | Bianca | Brian |" +
	"Camila | Carla | Carmen | Celine | Chantal | Conchita |" +
	" Cristiano | Dora | Emma | Enrique | Ewa | Filiz |" +
	" Geraint | Giorgio | Gwyneth | Hans | Ines | Ivy |" +
	" Jacek | Jan | Joanna | Joey | Justin | Karl | Kendra | Kevin |" +
	" Kimberly | Lea | Liv | Lotte | Lucia | Lupe | Mads | Maja |" +
	" Marlene | Mathieu | Matthew | Maxim | Mia | Miguel | Mizuki |" +
	" Naja | Nicole | Penelope | Raveena | Ricardo | Ruben | Russell |" +
	" Salli | Seoyeon | Takumi | Tatyana | Vicki | Vitoria | Zeina | Zhiyu"

func ProcessSpeak(msAdapter *adapters.MediaServer, data models.CallRequest, element *etree.Element) error {
	callSid := data.CallSid
	if data.Status != "in-progress" {
		_ = (*msAdapter).AnswerCall(data.CallSid)
		/*
			If its webrtc calls then send silence stream here for 2.5 second
		 */
		if data.SrcType == "Wss" || data.DestType == "Wss"{
			_ = (*msAdapter).PlayMediaFile(data.CallSid, "silence_stream://2000", "1")
		}
	}
	loopCount := 1
	voice := "default"
	language := "en-US"
	var err error
	for _, attr := range element.Attr {
		switch attr.Key {
		case "voice":
			voice = attr.Value
		case "loop":
			loopCount, err = strconv.Atoi(attr.Value)
			if err != nil {
				loopCount = 1
			}
			if loopCount == 0 {
				loopCount = 1000
			}
		case "language":
			language = attr.Value
		}
	}
	/*
		restClient := resty.New()
		resp, _ := restClient.R().
			SetBody(map[string]interface{}{"outputFormat": ttsOutputFormat,
				"sampleRate": ttsSampleRate,
				"inputText":  element.Text(), "voiceId": ttsVoiceId}).Post(ttsUrl)
		ttsResponse := make(map[string]interface{})
		json.Unmarshal(resp.Body(), &ttsResponse)
		ttsURL := ttsResponse["tts_file"].(string)
		logger.Logger.Debug("tts response uuid=% ttsURL=%s", ttsURL)
		strLoopCount := strconv.Itoa(loopCount)
		err = (*msAdapter).PlayMediaFile(uuid, ttsURL, strLoopCount)
		return err
	*/
	if !isValidAliceLanguage("language") {
		language = "en-US"
	}
	if !isValidVoice(voice) {
		voice = "default"
	}

	voiceText := strings.Replace(element.Text(), "\n", "", -1)
	logger.Logger.WithField("uuid", callSid).Info("voice is ", voice, " language is ", language)
	voiceId := getAwsVoiceId(voice, language)
	logger.Logger.WithField("uuid", callSid).Info("voice id is ", voiceId)
	err = (*msAdapter).Speak(callSid, voiceId, voiceText)
	if err != nil {
		return err
	}
	loopCount = loopCount - 1
	for loopCount > 0 {
		if err := (*msAdapter).Speak(callSid, voiceId, element.Text()); err != nil {
			return err
		}
		loopCount = loopCount - 1
	}
	return err
}

/*
	speak from rest
	{
		"to":"your_destination",
		"from":"your_callerId",
		"speak":"Welcome to tiniyo, We are here to help you"
	}
*/
func ProcessSpeakText(msAdapter *adapters.MediaServer, uuid string, speakText string) error {
	voiceId := "Salli"
	loopCount := 3
	var err error
	if err = (*msAdapter).Speak(uuid, voiceId, speakText); err != nil {
		return err
	}
	loopCount = loopCount - 1
	for loopCount > 0 {
		if err = (*msAdapter).Speak(uuid, voiceId, speakText); err != nil {
			return err
		}
		loopCount = loopCount - 1
	}
	return err
}

func isValidManLanguage(lang string) bool {
	switch lang {
	case "en-US", "en-GB",
		"es-ES", "fr-FR", "de-DE":
		return true
	}
	return false
}

func isValidAliceLanguage(lang string) bool {
	switch lang {
	case "en-US", "de-DE",
		"en-AU", "en-CA",
		"en-GB", "en-IN",
		"sv-SE", "zh-CN",
		"pl-PL", "pt-BR",
		"pt-PT", "ru-RU",
		"fr-CA", "fr-FR",
		"it-IT", "ja-JP",
		"ko-KR", "nb-NO",
		"nl-NL", "ca-ES",
		"es-ES", "es-MX",
		"fi-FI", "zh-HK",
		"zh-TW", "da-DK":
		return true
	}
	return false
}

func isValidVoice(voice string) bool {
	if strings.Contains(voice, "Polly") {
		voice = "Polly"
	}
	switch voice {
	case "man", "woman", "Polly", "alice", "default":
		return true
	}
	return false
}

func isValidLoop(loop int) bool {
	return false
}

func getAwsVoiceId(voice, lang string) string {
	voiceList := strings.SplitN(voice, ".", -1)
	if voiceList[0] == "Polly" {
		voiceId := voiceList[1]
		if strings.Contains(validAwsPollyVoiceId, voiceId) {
			return voiceId
		}
		voice = "default"
	}

	if lang == "" {
		lang = "en-US"
	}
	if voice == "" {
		voice = "default"
	}
	awsVoiceId["en-US"]["man"] = "Joey"
	awsVoiceId["en-US"]["woman"] = "Salli"
	awsVoiceId["en-US"]["alice"] = "Kendra"
	awsVoiceId["en-US"]["default"] = "Salli"

	awsVoiceId["de-DE"]["man"] = "Hans"
	awsVoiceId["de-DE"]["woman"] = "Marlene"
	awsVoiceId["de-DE"]["alice"] = "Vicki"
	awsVoiceId["de-DE"]["default"] = "Marlene"

	awsVoiceId["en-AU"]["man"] = "Russell"
	awsVoiceId["en-AU"]["woman"] = "Nicole"
	awsVoiceId["en-AU"]["alice"] = "Olivia"
	awsVoiceId["en-AU"]["default"] = "Nicole"

	awsVoiceId["en-GB"]["man"] = "Brian"
	awsVoiceId["en-GB"]["woman"] = "Emma"
	awsVoiceId["en-GB"]["alice"] = "Amy"
	awsVoiceId["en-GB"]["default"] = "Emma"

	awsVoiceId["en-IN"]["man"] = "Raveena"
	awsVoiceId["en-IN"]["woman"] = "Aditi"
	awsVoiceId["en-IN"]["alice"] = "Raveena"
	awsVoiceId["en-IN"]["default"] = "Raveena"

	awsVoiceId["da-DK"]["man"] = "Mads"
	awsVoiceId["da-DK"]["woman"] = "Naja"
	awsVoiceId["da-DK"]["alice"] = "Naja"
	awsVoiceId["da-DK"]["default"] = "Naja"

	awsVoiceId["nl-NL"]["man"] = "Ruben"
	awsVoiceId["nl-NL"]["woman"] = "Lotte"
	awsVoiceId["nl-NL"]["alice"] = "Lotte"
	awsVoiceId["nl-NL"]["default"] = "Lotte"

	awsVoiceId["es-MX"]["man"] = "Mia"
	awsVoiceId["es-MX"]["woman"] = "Mia"
	awsVoiceId["es-MX"]["alice"] = "Mia"
	awsVoiceId["es-MX"]["default"] = "Mia"

	awsVoiceId["sv-SE"]["man"] = "Astrid"
	awsVoiceId["sv-SE"]["woman"] = "Astrid"
	awsVoiceId["sv-SE"]["alice"] = "Astrid"
	awsVoiceId["sv-SE"]["default"] = "Astrid"

	awsVoiceId["pl-PL"]["man"] = "Jan"
	awsVoiceId["pl-PL"]["woman"] = "Ewa"
	awsVoiceId["pl-PL"]["alice"] = "Maja"
	awsVoiceId["pl-PL"]["default"] = "Ewa"

	awsVoiceId["pt-BR"]["man"] = "Camila"
	awsVoiceId["pt-BR"]["woman"] = "Camila"
	awsVoiceId["pt-BR"]["alice"] = "Camila"
	awsVoiceId["pt-BR"]["default"] = "Camila"

	awsVoiceId["ja-JP"]["man"] = "Takumi"
	awsVoiceId["ja-JP"]["woman"] = "Mizuki"
	awsVoiceId["ja-JP"]["alice"] = "Mizuki"
	awsVoiceId["ja-JP"]["default"] = "Mizuki"

	awsVoiceId["ko-KR"]["man"] = "Seoyeon"
	awsVoiceId["ko-KR"]["woman"] = "Seoyeon"
	awsVoiceId["ko-KR"]["alice"] = "Seoyeon"
	awsVoiceId["ko-KR"]["default"] = "Seoyeon"

	awsVoiceId["nb-NO"]["man"] = "Liv"
	awsVoiceId["nb-NO"]["woman"] = "Liv"
	awsVoiceId["nb-NO"]["alice"] = "Liv"
	awsVoiceId["nb-NO"]["default"] = "Liv"

	awsVoiceId["pt-PT"]["man"] = "Cristiano"
	awsVoiceId["pt-PT"]["woman"] = "Ines"
	awsVoiceId["pt-PT"]["alice"] = "Ines"
	awsVoiceId["pt-PT"]["default"] = "Ines"

	awsVoiceId["ru-RU"]["man"] = "Maxim"
	awsVoiceId["ru-RU"]["woman"] = "Tatyana"
	awsVoiceId["ru-RU"]["alice"] = "Tatyana"
	awsVoiceId["ru-RU"]["default"] = "Tatyana"

	awsVoiceId["fr-CA"]["man"] = "Chantal"
	awsVoiceId["fr-CA"]["woman"] = "Chantal"
	awsVoiceId["fr-CA"]["alice"] = "Chantal"
	awsVoiceId["fr-CA"]["default"] = "Chantal"

	awsVoiceId["fr-FR"]["man"] = "Mathieu"
	awsVoiceId["fr-FR"]["woman"] = "Celine"
	awsVoiceId["fr-FR"]["alice"] = "Celine"
	awsVoiceId["fr-FR"]["default"] = "Celine"

	awsVoiceId["it-IT"]["man"] = "Giorgio"
	awsVoiceId["it-IT"]["woman"] = "Carla"
	awsVoiceId["it-IT"]["alice"] = "Carla"
	awsVoiceId["it-IT"]["default"] = "Carla"

	awsVoiceId["es-ES"]["man"] = "Enrique"
	awsVoiceId["es-ES"]["woman"] = "Conchita"
	awsVoiceId["es-ES"]["alice"] = "Conchita"
	awsVoiceId["es-ES"]["default"] = "Conchita"

	awsVoiceId["en-CA"]["man"] = ""
	awsVoiceId["en-CA"]["woman"] = ""
	awsVoiceId["en-CA"]["alice"] = ""
	awsVoiceId["en-CA"]["default"] = ""

	awsVoiceId["zh-CN"]["man"] = ""
	awsVoiceId["zh-CN"]["woman"] = ""
	awsVoiceId["zh-CN"]["alice"] = ""
	awsVoiceId["zh-CN"]["default"] = ""

	awsVoiceId["ca-ES"]["man"] = ""
	awsVoiceId["ca-ES"]["woman"] = ""
	awsVoiceId["ca-ES"]["alice"] = ""
	awsVoiceId["ca-ES"]["default"] = ""

	awsVoiceId["fi-FI"]["man"] = ""
	awsVoiceId["fi-FI"]["woman"] = ""
	awsVoiceId["fi-FI"]["alice"] = ""
	awsVoiceId["fi-FI"]["default"] = ""

	awsVoiceId["zh-HK"]["man"] = ""
	awsVoiceId["zh-HK"]["woman"] = ""
	awsVoiceId["zh-HK"]["alice"] = ""
	awsVoiceId["zh-HK"]["default"] = ""

	awsVoiceId["zh-TW"]["man"] = ""
	awsVoiceId["zh-TW"]["woman"] = ""
	awsVoiceId["zh-TW"]["alice"] = ""
	awsVoiceId["zh-TW"]["default"] = ""

	if awsVoiceId[lang][voice] == "" {
		return "Salli"
	}
	return awsVoiceId[lang][voice]
}
