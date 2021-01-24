#include <Arduino.h>
#include <Adafruit_NeoPixel.h>

// Configurable values:
// Make sure that all the arrays are of the same length!
int pixelBus = 6;
int ledInputs[] = {2, 3, 4};
String ledNames[] = {"NUM", "SHIFT", "FLOCK"};
uint32_t ledColors[] = {0, 0, 0};
uint32_t overrideColors[] = {0, 0, 0};
bool ledForced[] = {false, false, false};

int pixelFormat = NEO_GRB + NEO_KHZ800;

Adafruit_NeoPixel *pixels;

int numPixels = 0;

void setup() {
  Serial.begin(9600);

  numPixels = sizeof(ledInputs) / sizeof(int);
  if (numPixels != sizeof(ledColors) / sizeof(uint32_t) || numPixels != sizeof(ledForced) / sizeof(bool) || numPixels != sizeof(ledNames) / sizeof(String)) {
    Serial.println("ledInputs, ledColors, ledNames and ledForced have to be of the same length");
  }

  Serial.setTimeout(30000);

  pixels = new Adafruit_NeoPixel(numPixels, pixelBus, pixelFormat);
  pixels->begin();

  

  for (int i=0; i<numPixels; i++) {
    pinMode(ledInputs[i], INPUT_PULLUP);
    overrideColors[i] = pixels->Color(0, 0, 150);
  }
}

void loop() {
  if (Serial.available() > 0) {
    String incomingComand = Serial.readStringUntil('\n');
    bool ok = false;
    String status = "";
    String ledName = "";

    for (int i=0; i<numPixels; i++) {
      if (!incomingComand.startsWith(ledNames[i])) {
        // Found no led with that name.
        if (i == numPixels-1) {
          status = "invalid led id";
        }
        continue;
      }

      ledName = ledNames[i];

      incomingComand = incomingComand.substring(ledNames[i].length() + 1);
      Serial.println(incomingComand);
      if (incomingComand.startsWith("disable override")) {
        ledForced[i] = true;
        status = "override disabled";
        ok = true;
        break;
      } else {
        bool setOverrideColor = false;
        if (incomingComand.startsWith("override")) {
          incomingComand = incomingComand.substring(9);
          setOverrideColor = true;
          ledForced[i] = false;
        }

        // Parse the color
        String r = incomingComand.substring(0, 2);
        String g = incomingComand.substring(2, 4);
        String b = incomingComand.substring(4, 6);

        uint32_t newColor = pixels->Color(
          (uint8_t) strtol(r.c_str(), 0, 16),
          (uint8_t) strtol(g.c_str(), 0, 16),
          (uint8_t) strtol(b.c_str(), 0, 16));

        if (setOverrideColor) {
          overrideColors[i] = newColor;
          status = "override color set";
        } else {
          ledColors[i] = newColor;
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

  for (int i=0; i<numPixels; i++) {
    pixels->setPixelColor(i, ledColors[i]);

    if (digitalRead(ledInputs[i]) == LOW && !ledForced[i]) {
      pixels->setPixelColor(i, overrideColors[i]);
    }
  }

  pixels->show();
}