# Lantern on Android

The `lantern-android` repository provides documentation and scripts for
building a basic [flashlight][1] client library for Android devices.

## Prerequisites

* An OSX or Linux box
* [docker][2]
* [Android Studio][3]
* [Go][4]
* [GNUMake][6]

### Setting up the experimental environment

This is an experimental feature so we need to do some minor hacks in order to
test it. We're going to work with the `experimental/lantern-android` branch of
[flashlight-build][5]:

```sh
mkdir -p $GOPATH/src/github.com/getlantern
cd $GOPATH/src/github.com/getlantern
git clone https://github.com/getlantern/flashlight-build.git
cd flashlight-build
git checkout -b experimental/lantern-android remotes/origin/experimental/lantern-android
```

Please note this is only a temporary hack while we wait for the required
changes to hit upstream.

## Building the Android library

Set the `GOPATH` environment variable to
`$GOPATH/src/github.com/getlantern/flashlight-build` for the current session,
the [flashlight-build][5] repository has everything we need to build the
[flashlight][1] lightweight web proxy:

```sh
export GOPATH=$GOPATH/src/github.com/getlantern/flashlight-build
```

Change directory to `$GOPATH/src/github.com/getlantern` and clone the
`lantern-android` repository with git.

```sh
cd $GOPATH/src/github.com/getlantern
git clone https://github.com/getlantern/lantern-android.git
```

Finally, change directory into `lantern-android` and pass the build task to the
`make` command.

```
make
# ...
# BUILD SUCCESSFUL
# Total time: 7 seconds
```

The `make` command will create a new `app` subdirectory that will contain an
Android example project. You may import the contents of the `app` subdirectory
into Android Studio to see libflashlight working.

## Testing the example project

Open [Android Studio][3] and in the welcome screen choose "Import Non-Android
Studio project".

![Android Studio](https://cloud.githubusercontent.com/assets/385670/5712830/5f4cda3c-9a7b-11e4-85af-8af9d54e18c7.png)

You'll be prompted with a file dialog, browse to the `app` subdirectory and
select it. Press *OK*.

![App Subdirectory](https://cloud.githubusercontent.com/assets/385670/5712859/91454204-9a7b-11e4-8227-e1372ff65ee4.png)

On the next dialog you must define a destination for the project, hit *Next*.

![Destination](https://cloud.githubusercontent.com/assets/385670/5712874/ad8265e6-9a7b-11e4-9018-671875dfdb17.png)

Add a new *main activity* by right-clicking on the top most directory on the
*Project* pane and selecting New->Activity->Blank Activity, the default values
would be OK, click *Finish*.

![Main Activity](https://cloud.githubusercontent.com/assets/385670/5712891/ca3573fe-9a7b-11e4-953d-d43b12fcdb62.png)

Paste the following code on the `org.getlantern/example/MainActivity.java` file
that was just added:

```java
package org.getlantern.example;

import go.Go;
import go.flashlight.Flashlight;
import android.app.Activity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import org.getlantern.example.R;


public class MainActivity extends Activity {

    private Thread proxyThread = null;
    private Runnable proxyRunnable = null;
    private Button killButton;
    private Button startButton;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        setContentView(R.layout.activity_main);

        // Initializing application context.
        Go.init(getApplicationContext());

        killButton = (Button)findViewById(R.id.stopProxyButton);
        startButton = (Button)findViewById(R.id.startProxyButton);

        // Disabling stop button.
        killButton.setEnabled(false);

        // Enabling proxy button.
        startButton.setEnabled(true);
    }

    public void stopProxyButtonOnClick(View v) {
        if (proxyThread.isAlive()) {
            Log.v("DEBUG", "Attempt to stop running proxy.");
            Flashlight.StopClientProxy();

            // Disabling stop button.
            killButton.setEnabled(false);

            // Enabling proxy button.
            startButton.setEnabled(true);

            // Asking the thread to interrupt itself.
            proxyThread.interrupt();
            proxyThread = null;
        }
    }

    public void startProxyButtonOnClick(View v) {
        Log.v("DEBUG", "Attempt to run client proxy on :9192");

        if (proxyThread == null || proxyThread.isInterrupted()) {
            proxyRunnable = new Runnable() {
                @Override
                public void run() {
                    try {
                        Log.v("DEBUG", "Running proxy...");
                        Flashlight.RunClientProxy("0.0.0.0:9192");
                    } catch (Exception e) {
                        throw new RuntimeException(e);
                    }
                }
            };

            // Creating a new thread so the UI does not block.
            proxyThread = new Thread(proxyRunnable);

            proxyThread.start();

        } else {
            Log.v("DEBUG", "Proxy is already running");
        }

        // Enabling stop button.
        killButton.setEnabled(true);

        // Disabling proxy button.
        startButton.setEnabled(false);
    }
}
```

After this new activity is added the *design view* will be active, drag two
buttons from the *Pallete* into the screen.

![Adding two buttons](https://cloud.githubusercontent.com/assets/385670/5769192/d9df19cc-9cdd-11e4-90d0-b37b6d6b3a41.png)

Select the first button and look for the *id* property on the Properties pane,
set it to *startProxyButton* and name the button accordingly. Look for the
*onClick* property and choose the *startProxyButtonOnClick* value from the drop
down.

The second button's *id* must be set to *stopProxyButton* and the *onClick* to
*stopProxyButtonOnClick*.

Finally, hit the *Run app* action under the *Run* menu and deploy it to a real
device or to an ARM-based emulator.

If everything goes OK, you'll have two buttons and you can start `flashlight`
by touching the *startProxyButton*.

![Deploy to a device](https://cloud.githubusercontent.com/assets/385670/5712899/db6ddb34-9a7b-11e4-8841-6b6b12e46c27.png)

As long as the app is open, you'll be able to test the canonical example by
finding the device's IP and sending it a special request:

```
curl -x 10.10.100.97:9192 http://www.google.com/humans.txt
# Google is built by a large team of engineers, designers, researchers, robots, and others in many different sites across the globe. It is updated continuously, and built with more tools and technologies than we can shake a stick at. If you'd like to help us out, see google.com/careers.
```

You may not want everyone proxying through your phone! Tune the
`RunClientProxy()` function on the `MainActivity.java` accordingly.

Note: The *stopProxyButton* is supposed to stop the flashlight proxy but this
is currently not supported, so at this time *stopProxyButton* actually kills
the app.

## Building a stand-alone client binary for Android devices

(pending)

[1]: https://github.com/getlantern/flashlight
[2]: https://www.docker.com/
[3]: http://developer.android.com/tools/studio/index.html
[4]: http://golang.org/
[5]: https://github.com/getlantern/flashlight-build
[6]: http://www.gnu.org/software/make/
