#include <Arduino.h>
#include <Adafruit_NeoPixel.h>
#include <ArduinoJson.h>

struct LedData {
  int input;
  String id;
  uint32_t color;
  uint32_t overrideColor;
  bool overrideInverted;
  bool isForced;
};

// Configurable values:
// Make sure that all the arrays are of the same length!
int pixelBus = 6;
float brightnessScale = 0.6;

LedData ledConfig[] = {
  LedData{
    input: 3,
    id: "FLOCK",
    color: 0,
    overrideColor: 150,
    overrideInverted: true,
    isForced: false
  },
  LedData{
    input: 2,
    id: "SHIFT",
    color: 0,
    overrideColor:  150,
    overrideInverted: false,
    isForced: false
  },
  LedData{
    input: 4,
    id: "NUM",
    color: 0,
    overrideColor:  150,
    overrideInverted: false,
    isForced: false
  }
};

int pixelFormat = NEO_GRB + NEO_KHZ800;

Adafruit_NeoPixel *pixels;

int numLeds = 0;

void setup() {
  Serial.begin(9600);

  numLeds = sizeof(ledConfig) / sizeof(LedData);

  Serial.setTimeout(30000);

  pixels = new Adafruit_NeoPixel(numLeds, pixelBus, pixelFormat);
  pixels->begin();

  for (int i=0; i<numLeds; i++) {
    pinMode(ledConfig[i].input, INPUT_PULLUP);
  }

  Serial.println("Started...");
}

DynamicJsonDocument doc(1024);

void loop() {
  if (Serial.available() > 0) {
    String incomingComand = Serial.readStringUntil('\n');
    bool ok = false;
    String status = "";
    String ledName = "";

    // Send own capabilities as json.
    if (incomingComand.startsWith("hello")) {
      doc.clear();

      doc["name"] = "keyboard-mod";
      doc["version"] = "0.1.0";
      JsonArray leds = doc.createNestedArray("leds");
      for (int i=0; i<numLeds; i++) {
        JsonObject led = leds.createNestedObject();
        led["id"] = ledConfig[i].id;
      }
      Serial.print("OK: hello ");
      serializeJson(doc, Serial);
      Serial.println();
      Serial.flush();
      return;
    }

    for (int i=0; i<numLeds; i++) {
      status = "";
      if (!incomingComand.startsWith(ledConfig[i].id)) {
        // Found no led with that name.
        if (i == numLeds-1) {
          status = "invalid led id";
        }
        continue;
      }

      ledName = ledConfig[i].id;

      incomingComand = incomingComand.substring(ledConfig[i].id.length() + 1);
      if (incomingComand.startsWith("disable override")) {
        ledConfig[i].isForced = true;
        status = "override disabled";
        ok = true;
        break;
      } else {
        bool setOverrideColor = false;
        if (incomingComand.startsWith("override")) {
          incomingComand = incomingComand.substring(9);
          setOverrideColor = true;
          ledConfig[i].isForced = false;
        }

        // Parse the color
        String r = incomingComand.substring(0, 2);
        String g = incomingComand.substring(2, 4);
        String b = incomingComand.substring(4, 6);

        uint32_t newColor = pixels->Color(
          (uint8_t) strtol(r.c_str(), 0, 16) * brightnessScale,
          (uint8_t) strtol(g.c_str(), 0, 16) * brightnessScale,
          (uint8_t) strtol(b.c_str(), 0, 16) * brightnessScale);

        if (setOverrideColor) {
          ledConfig[i].overrideColor = newColor;
          status = "override color set";
        } else {
          ledConfig[i].color = newColor;
          status = "color set";
        }

        ok = true;
        break;
      }
    }

    if (ok) {
      doc.clear();
      doc["id"] = ledName;
      doc["status"] = status;
      
      Serial.print("OK: set ");
      serializeJson(doc, Serial);
      Serial.println();
      Serial.flush();
    } else  {
      doc.clear();
      doc["id"] = ledName;
      doc["status"] = status;
      
      Serial.print("ERR: ");
      serializeJson(doc, Serial);
      Serial.println();
      Serial.flush();
    }
  }

  pixels->clear();

  for (int i=0; i<numLeds; i++) {
    uint32_t newColor = ledConfig[i].color;

    bool input = digitalRead(ledConfig[i].input);
    if (ledConfig[i].overrideInverted) {
      input = !input;
    }

    if (input == LOW && !ledConfig[i].isForced) {
      newColor = ledConfig[i].overrideColor;
    }

    pixels->setPixelColor(i, newColor);
  }

  pixels->show();
}