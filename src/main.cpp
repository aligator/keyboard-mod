#include <Arduino.h>
#include <Adafruit_NeoPixel.h>

struct LedData {
  int input;
  String name;
  uint32_t color;
  uint32_t overrideColor;
  bool overrideInverted;
  bool isForced;
};

// Configurable values:
// Make sure that all the arrays are of the same length!
int pixelBus = 6;
float brightnessScale = 0.6;

LedData leds[] = {
  LedData{
    input: 3,
    name: "FLOCK",
    color: 0,
    overrideColor: 150,
    overrideInverted: true,
    isForced: false
  },
  LedData{
    input: 2,
    name: "SHIFT",
    color: 0,
    overrideColor:  150,
    overrideInverted: false,
    isForced: false
  },
  LedData{
    input: 4,
    name: "NUM",
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

  numLeds = sizeof(leds) / sizeof(LedData);

  Serial.setTimeout(30000);

  pixels = new Adafruit_NeoPixel(numLeds, pixelBus, pixelFormat);
  pixels->begin();

  for (int i=0; i<numLeds; i++) {
    pinMode(leds[i].input, INPUT_PULLUP);
  }

  Serial.println("Started...");
}

void loop() {
  if (Serial.available() > 0) {
    String incomingComand = Serial.readStringUntil('\n');
    bool ok = false;
    String status = "";
    String ledName = "";

    for (int i=0; i<numLeds; i++) {
      if (!incomingComand.startsWith(leds[i].name)) {
        // Found no led with that name.
        if (i == numLeds-1) {
          status = "invalid led id";
        }
        continue;
      }

      ledName = leds[i].name;

      incomingComand = incomingComand.substring(leds[i].name.length() + 1);
      Serial.println(incomingComand);
      if (incomingComand.startsWith("disable override")) {
        leds[i].isForced = true;
        status = "override disabled";
        ok = true;
        break;
      } else {
        bool setOverrideColor = false;
        if (incomingComand.startsWith("override")) {
          incomingComand = incomingComand.substring(9);
          setOverrideColor = true;
          leds[i].isForced = false;
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
          leds[i].overrideColor = newColor;
          status = "override color set";
        } else {
          leds[i].color = newColor;
          status = "color set";
        }

        ok = true;
        break;
      }
    }

    if (ok) {
      Serial.println("OK: " + ledName + " " + status);
    } else  {
      Serial.println("ERR: " + status);
    }
  }

  pixels->clear();

  for (int i=0; i<numLeds; i++) {
    uint32_t newColor = leds[i].color;

    Serial.println(leds[i].name);
    Serial.println(i);
    Serial.println(leds[i].input);
    Serial.println(leds[i].overrideInverted);

    bool input = digitalRead(leds[i].input);
    if (leds[i].overrideInverted) {
      input = !input;
    }

    if (input == LOW && !leds[i].isForced) {
      newColor = leds[i].overrideColor;
    }

    Serial.println(newColor);


    pixels->setPixelColor(i, newColor);
  }

  pixels->show();
}