# SyncRide End-to-End (E2E) Physical Device Testing Guide

This guide provides step-by-step instructions for testing the **SyncRide** native Android application on a physical mobile device connected to your local backend development infrastructure (gRPC, PostgreSQL, Kafka, and Go microservices).

---

## 1. Network & IP Configuration

To enable your physical Android phone to communicate with the microservices running on your host PC, both devices must be on the same local area network (LAN).

### 1.1 Wi-Fi Network Setup
- Ensure your **physical phone** and **PC** are connected to the **exact same Wi-Fi network**.
- Cellular data on the phone should be disabled during local testing to prevent routing issues.

### 1.2 Find your PC's Local IPv4 Address
Open **Command Prompt** or **PowerShell** on your PC and run:

```cmd
ipconfig
```

Look for your active network adapter (e.g., *Wireless LAN adapter Wi-Fi* or *Ethernet*) and copy the **IPv4 Address**:

```text
Wireless LAN adapter Wi-Fi:
   Connection-specific DNS Suffix  . : localdomain
   IPv4 Address. . . . . . . . . . . : 192.168.1.50   <-- Copy this IP
   Subnet Mask . . . . . . . . . . . : 255.255.255.0
   Default Gateway . . . . . . . . . : 192.168.1.1
```

> **Note:** Write down your local IP (e.g., `192.168.1.50`). You will use this IP instead of `localhost` or `10.0.2.2`.

### 1.3 Windows Defender Firewall Setup

> ⚠️ **CRITICAL WARNING:** Windows Defender Firewall blocks incoming connections on non-standard ports by default. You MUST allow incoming traffic on ports `:8080` (API Gateway) and `:50051` (Ride Service gRPC) for your phone to connect.

#### How to Allow Ports in Windows Firewall:
1. Press `Win + R`, type `wf.msc`, and press **Enter** to open *Windows Defender Firewall with Advanced Security*.
2. In the left panel, click **Inbound Rules**.
3. In the right panel, click **New Rule...**.
4. Select **Port** and click **Next**.
5. Select **TCP** and enter **Specific local ports**: `8080, 50051`, then click **Next**.
6. Select **Allow the connection** and click **Next**.
7. Keep all profiles checked (**Domain**, **Private**, **Public**) and click **Next**.
8. Name the rule `SyncRide Backend Local Access` and click **Finish**.

---

## 2. Android App Code Modification

Before flashing the app to your physical device, update the networking layer to point to your PC's LAN IP address.

### 2.1 Update Repository Code
Pull the latest Android codebase from your GitHub repository onto your PC:

```bash
git pull origin main
```

### 2.2 Configure Base URL
Locate the network configuration file in the Android project (e.g., `NetworkModule.kt`, `Constants.kt`, or `RetrofitClient.kt` under `app/src/main/java/...`):

Replace any reference to `localhost`, `127.0.0.1`, or `10.0.2.2` with your PC's IPv4 address:

```kotlin
// Example in NetworkModule.kt / Constants.kt
object NetworkConstants {
    // Replace 192.168.1.50 with your actual PC IPv4 address from Section 1.2
    const val API_GATEWAY_BASE_URL = "http://192.168.1.50:8080/"
    const val RIDE_SERVICE_GRPC_HOST = "192.168.1.50"
    const val RIDE_SERVICE_GRPC_PORT = 50051
}
```

### 2.3 Commit & Push or Build Locally
You can either commit and push the change to your feature branch or compile directly in Android Studio.

---

## 3. Deployment to Physical Device

### 3.1 Enable Developer Options on Phone
1. Open **Settings** on your Android phone.
2. Navigate to **About Phone** -> **Software Information**.
3. Tap **Build Number** seven (7) times until you see the message *"You are now a developer!"*.
4. Return to **Settings** -> **Developer Options**.
5. Enable **USB Debugging** (and optionally **Wireless Debugging** if on Android 11+).

### 3.2 Flash App from Android Studio
1. Open the `SyncRide` Android project in **Android Studio**.
2. Connect your phone to your PC via USB cable.
3. Accept the **"Allow USB Debugging?"** prompt on your phone screen.
4. Select your physical phone model from the target device dropdown menu in Android Studio's top toolbar.
5. Click **Run 'app'** (`Shift + F10`) or **Debug 'app'**.

*Alternative CLI Installation via Gradle & ADB:*
```bash
./gradlew assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

---

## 4. Execution & Log Verification (The Test)

Follow these 4 steps to execute the end-to-end flow from phone to backend:

### Step 1: Launch Backend Services
On your PC, navigate to the `backend/` directory and run the automation script:

```cmd
run_syncride.bat
```

This starts PostgreSQL, Kafka, and launches the `ride-service` gRPC server listening on port `:50051`.

### Step 2: Trigger Ride Request from Phone
1. Open the **SyncRide** app on your physical Android phone.
2. Log in or select a passenger profile.
3. Set your pickup location (e.g., Jaipur Hawa Mahal) and destination.
4. Tap **Request Ride**.

### Step 3: Monitor PC Terminal Logs
Watch the **Ride Service** terminal window on your PC (`[Ride Service]`). You will observe real-time log output confirming receipt of the request:

```text
Starting Ride Service...
PostgreSQL connection established and migration completed successfully.
Kafka producer configured with broker localhost:9092
Ride Service gRPC server listening on :50051
[gRPC] Received RequestRide RPC from passenger_id: e3b0c442-98fc-11ee-b9d1-0242ac120002
[Kafka] Published event 'ride.requested' to topic 'ride.requested'
```

### Step 4: Verify Event in Kafka UI
1. Open your web browser on your PC and navigate to **Kafka UI**:
   [http://localhost:8081](http://localhost:8081)
2. In the left navigation bar, select **Topics**.
3. Click on the `ride.requested` topic.
4. Click on the **Messages** tab.
5. Verify that the new event payload dispatched from your physical phone appears with the matching `passenger_id`, `pickup_lat`, `pickup_lng`, and `ride_type`:

```json
{
  "ride_id": "3c7e6e8c-4c47-44e7-8359-06fbccc3cbc9",
  "passenger_id": "e3b0c442-98fc-11ee-b9d1-0242ac120002",
  "pickup_lat": 26.923932,
  "pickup_lng": 75.826744,
  "dropoff_lat": 26.853043,
  "dropoff_lng": 75.804677,
  "requested_at": "2026-08-09T16:25:58Z"
}
```

---

## 5. Troubleshooting Summary

| Symptom | Cause | Solution |
| :--- | :--- | :--- |
| `java.net.ConnectException: Connection refused` | Phone pointing to `localhost` or wrong IP | Update `NetworkConstants` with PC IPv4 address from `ipconfig`. |
| `gRPC Unavailable / Connection Timeout` | Windows Firewall blocking port `:50051` | Add Inbound Firewall rule for ports `8080` and `50051`. |
| `CLEARTEXT communication not permitted` | Android blocking HTTP traffic | Ensure `android:usesCleartextTraffic="true"` is set in `AndroidManifest.xml`. |
