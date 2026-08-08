package com.syncride.app

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Color
import android.graphics.Typeface
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.View
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.cardview.widget.CardView
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import org.osmdroid.config.Configuration
import org.osmdroid.util.BoundingBox
import org.osmdroid.util.GeoPoint
import org.osmdroid.views.MapView
import org.osmdroid.tileprovider.tilesource.TileSourceFactory
import org.osmdroid.views.overlay.Marker
import org.osmdroid.views.overlay.Polyline
import org.osmdroid.views.overlay.MapEventsOverlay
import org.osmdroid.events.MapEventsReceiver
import org.osmdroid.views.overlay.mylocation.GpsMyLocationProvider
import org.osmdroid.views.overlay.mylocation.MyLocationNewOverlay

class MainActivity : AppCompatActivity() {

    private lateinit var mapView: MapView
    private val LOCATION_REQUEST_CODE = 1

    private var selectedTransitMode: String = "Shared Cab"
    private var selectedTransitIcon: String = "🚗"
    private var selectedDestination: String? = null
    private var pickupLocation: String? = null
    
    private var currentRideFare: Int = 0
    private var isMapPickerMode: Boolean = false
    private var pickedPoint: GeoPoint? = null

    private val routeLauncher = registerForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
        if (result.resultCode == Activity.RESULT_OK) {
            val destination = result.data?.getStringExtra("SELECTED_DESTINATION")
            val pickup = result.data?.getStringExtra("PICKUP_LOCATION")
            
            if (destination == "MAP_PICKER") {
                enterMapPickerMode()
            } else if (destination != null) {
                selectedDestination = destination
                pickupLocation = pickup ?: "Current Location"
                showRouteOnMap(pickupLocation!!, selectedDestination!!)
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        Configuration.getInstance().userAgentValue = packageName
        setContentView(R.layout.activity_main)

        // 1. Setup Map
        mapView = findViewById(R.id.mapView)
        mapView.setTileSource(TileSourceFactory.MAPNIK)
        mapView.setMultiTouchControls(true)
        mapView.controller.setZoom(16.0)

        setupMapEvents()

        // 2. Initial State: Only Search card visible
        val searchCard = findViewById<CardView>(R.id.searchCard)
        searchCard.setOnClickListener {
            val intent = Intent(this, RouteActivity::class.java)
            routeLauncher.launch(intent)
        }

        // 3. Setup Ride Selectors (within the module card)
        setupTransitSelectors()

        // 4. GPS Permission Check
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.ACCESS_FINE_LOCATION), LOCATION_REQUEST_CODE)
        } else {
            enableUserLocation()
        }

        // 5. Bottom Navigation
        val navRide = findViewById<LinearLayout>(R.id.navRide)
        val navAllServices = findViewById<LinearLayout>(R.id.navAllServices)
        val navProfile = findViewById<LinearLayout>(R.id.navProfile)

        navRide?.setOnClickListener {
            updateBottomNav("Ride")
            resetMainUI()
        }
        navAllServices?.setOnClickListener {
            updateBottomNav("Services")
            startActivity(Intent(this, AllServicesActivity::class.java))
        }
        navProfile?.setOnClickListener {
            updateBottomNav("Profile")
            startActivity(Intent(this, ProfileActivity::class.java))
        }

        // 6. Confirm Map Location Button
        findViewById<Button>(R.id.btnConfirmMapLocation).setOnClickListener {
            if (pickedPoint != null) {
                exitMapPickerMode(true)
            }
        }

        // 7. Booking & Completion Buttons
        findViewById<Button>(R.id.btnConfirmRide).setOnClickListener {
            startMockBooking()
        }
        findViewById<TextView>(R.id.btnCancelSearch).setOnClickListener {
            resetMainUI()
        }
        findViewById<Button>(R.id.btnEndRide).setOnClickListener {
            completeRide()
        }

        // Initial setup of markers
        addMockVehicleMarkers()
    }

    private fun showRouteOnMap(pickup: String, destination: String) {
        findViewById<CardView>(R.id.searchCard).visibility = View.GONE
        findViewById<CardView>(R.id.rideModuleCard).visibility = View.VISIBLE
        
        mapView.overlays.removeAll { it is Polyline || it is Marker && it.title != "SyncRide Partner" }
        
        val startPoint = GeoPoint(24.5854, 73.7125) // Mock user/pickup
        val endPoint = GeoPoint(24.5980, 73.7250) // Mock destination
        
        val startMarker = Marker(mapView)
        startMarker.position = startPoint
        startMarker.title = "Pickup: $pickup"
        startMarker.setAnchor(Marker.ANCHOR_CENTER, Marker.ANCHOR_BOTTOM)
        mapView.overlays.add(startMarker)
        
        val endMarker = Marker(mapView)
        endMarker.position = endPoint
        endMarker.title = "Destination: $destination"
        endMarker.setAnchor(Marker.ANCHOR_CENTER, Marker.ANCHOR_BOTTOM)
        mapView.overlays.add(endMarker)

        val line = Polyline()
        line.setPoints(listOf(startPoint, endPoint))
        line.color = Color.parseColor("#3B82F6")
        line.width = 12f
        mapView.overlays.add(line)

        focusOnRoute(startPoint, endPoint)
        mapView.invalidate()
    }

    private fun focusOnRoute(p1: GeoPoint, p2: GeoPoint) {
        val box = BoundingBox.fromGeoPoints(listOf(p1, p2))
        mapView.zoomToBoundingBox(box.increaseByScale(1.5f), true)
    }

    private fun setupTransitSelectors() {
        val updateDetails = { mode: String, icon: String, fare: Int ->
            selectedTransitMode = mode
            selectedTransitIcon = icon
            currentRideFare = fare
            showRideBookingCard()
        }

        findViewById<LinearLayout>(R.id.cardSharedCab).setOnClickListener { updateDetails("Shared Cab", "🚗", 120) }
        findViewById<LinearLayout>(R.id.cardTaxi).setOnClickListener { updateDetails("Private Taxi", "🚕", 350) }
        findViewById<LinearLayout>(R.id.cardSharedAuto).setOnClickListener { updateDetails("Shared Auto", "🛺", 40) }
        findViewById<LinearLayout>(R.id.cardCityBus).setOnClickListener { updateDetails("City Bus", "🚌", 15) }
        findViewById<LinearLayout>(R.id.cardTraveller).setOnClickListener { updateDetails("Traveller", "🚐", 80) }
    }

    private fun showRideBookingCard() {
        findViewById<CardView>(R.id.rideModuleCard).visibility = View.GONE
        val detailsCard = findViewById<CardView>(R.id.rideDetailsCard)
        
        findViewById<TextView>(R.id.detailsDestination).text = selectedDestination
        findViewById<TextView>(R.id.detailsTransitIcon).text = selectedTransitIcon
        findViewById<TextView>(R.id.detailsTransitType).text = "$selectedTransitMode"
        findViewById<TextView>(R.id.detailsFare).text = "₹$currentRideFare"
        
        if (selectedTransitMode == "City Bus") {
            findViewById<Button>(R.id.btnConfirmRide).text = "Buy Transit Pass"
        } else {
            findViewById<Button>(R.id.btnConfirmRide).text = "Confirm Shared Ride"
        }
        
        detailsCard.visibility = View.VISIBLE
    }

    private fun startMockBooking() {
        findViewById<CardView>(R.id.rideDetailsCard).visibility = View.GONE
        findViewById<CardView>(R.id.searchingOverlay).visibility = View.VISIBLE
        
        Handler(Looper.getMainLooper()).postDelayed({
            findViewById<CardView>(R.id.searchingOverlay).visibility = View.GONE
            findViewById<CardView>(R.id.activeRideCard).visibility = View.VISIBLE
            animateVehicleArrival()
            Toast.makeText(this, "Success! Your ride is on the way.", Toast.LENGTH_LONG).show()
        }, 2500)
    }

    private fun animateVehicleArrival() {
        val driverMarker = Marker(mapView)
        val start = GeoPoint(24.5754, 73.7025)
        val userPos = GeoPoint(24.5854, 73.7125)
        
        driverMarker.position = start
        driverMarker.title = "Driver"
        mapView.overlays.add(driverMarker)

        val handler = Handler(Looper.getMainLooper())
        var progress = 0f
        val runnable = object : Runnable {
            override fun run() {
                progress += 0.04f
                if (progress <= 1.0f) {
                    val lat = start.latitude + (userPos.latitude - start.latitude) * progress
                    val lon = start.longitude + (userPos.longitude - start.longitude) * progress
                    driverMarker.position = GeoPoint(lat, lon)
                    mapView.invalidate()
                    handler.postDelayed(this, 100)
                }
            }
        }
        handler.post(runnable)
    }

    private fun completeRide() {
        val sharedPref = getSharedPreferences("SyncRidePrefs", Context.MODE_PRIVATE)
        val currentBalance = sharedPref.getFloat("WALLET_BALANCE", 0f)
        sharedPref.edit().apply {
            putFloat("WALLET_BALANCE", currentBalance - currentRideFare)
            putInt("RIDES_COUNT", sharedPref.getInt("RIDES_COUNT", 128) + 1)
            apply()
        }
        Toast.makeText(this, "Ride Complete. ₹$currentRideFare paid.", Toast.LENGTH_LONG).show()
        resetMainUI()
    }

    private fun resetMainUI() {
        findViewById<CardView>(R.id.searchCard).visibility = View.VISIBLE
        findViewById<CardView>(R.id.rideModuleCard).visibility = View.GONE
        findViewById<CardView>(R.id.rideDetailsCard).visibility = View.GONE
        findViewById<CardView>(R.id.searchingOverlay).visibility = View.GONE
        findViewById<CardView>(R.id.activeRideCard).visibility = View.GONE
        findViewById<Button>(R.id.btnConfirmMapLocation).visibility = View.GONE
        
        mapView.overlays.clear()
        setupMapEvents()
        enableUserLocation()
        addMockVehicleMarkers()
        mapView.invalidate()
        
        selectedDestination = null
        pickupLocation = null
        isMapPickerMode = false
    }

    private fun setupMapEvents() {
        val receiver = object : MapEventsReceiver {
            override fun singleTapConfirmedHelper(p: GeoPoint?): Boolean {
                if (isMapPickerMode && p != null) {
                    pickedPoint = p
                    updatePickedMarker(p)
                    return true
                }
                return false
            }
            override fun longPressHelper(p: GeoPoint?): Boolean = false
        }
        mapView.overlays.add(MapEventsOverlay(receiver))
    }

    private fun updatePickedMarker(p: GeoPoint) {
        mapView.overlays.removeAll { it is Marker && it.title == "Picked Location" }
        val marker = Marker(mapView)
        marker.position = p
        marker.title = "Picked Location"
        mapView.overlays.add(marker)
        mapView.invalidate()
    }

    private fun enterMapPickerMode() {
        isMapPickerMode = true
        findViewById<CardView>(R.id.searchCard).visibility = View.GONE
        findViewById<Button>(R.id.btnConfirmMapLocation).visibility = View.VISIBLE
        Toast.makeText(this, "Tap to set destination", Toast.LENGTH_SHORT).show()
    }

    private fun exitMapPickerMode(confirmed: Boolean) {
        isMapPickerMode = false
        findViewById<Button>(R.id.btnConfirmMapLocation).visibility = View.GONE
        if (confirmed && pickedPoint != null) {
            selectedDestination = "Map Location (${"%.4f".format(pickedPoint!!.latitude)}, ${"%.4f".format(pickedPoint!!.longitude)})"
            pickupLocation = "Current Location"
            showRouteOnMap(pickupLocation!!, selectedDestination!!)
        } else {
            resetMainUI()
        }
    }

    private fun updateBottomNav(activeTab: String) {
        val textRide = findViewById<TextView>(R.id.textRide)
        val textServices = findViewById<TextView>(R.id.textServices)
        val textProfile = findViewById<TextView>(R.id.textProfile)
        textRide.setTextColor(if (activeTab == "Ride") Color.parseColor("#111827") else Color.parseColor("#6B7280"))
        textServices.setTextColor(if (activeTab == "Services") Color.parseColor("#111827") else Color.parseColor("#6B7280"))
        textProfile.setTextColor(if (activeTab == "Profile") Color.parseColor("#111827") else Color.parseColor("#6B7280"))
        textRide.setTypeface(null, if (activeTab == "Ride") Typeface.BOLD else Typeface.NORMAL)
        textServices.setTypeface(null, if (activeTab == "Services") Typeface.BOLD else Typeface.NORMAL)
        textProfile.setTypeface(null, if (activeTab == "Profile") Typeface.BOLD else Typeface.NORMAL)
    }

    private fun addMockVehicleMarkers() {
        val points = listOf(GeoPoint(24.5880, 73.7150), GeoPoint(24.5820, 73.7100), GeoPoint(24.5900, 73.7080))
        points.forEach { geoPoint ->
            val marker = Marker(mapView)
            marker.position = geoPoint
            marker.title = "SyncRide Partner"
            mapView.overlays.add(marker)
        }
    }

    private fun enableUserLocation() {
        val locationOverlay = MyLocationNewOverlay(GpsMyLocationProvider(this), mapView)
        locationOverlay.enableMyLocation()
        locationOverlay.enableFollowLocation()
        mapView.overlays.add(locationOverlay)
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == LOCATION_REQUEST_CODE && grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
            enableUserLocation()
        }
    }

    override fun onResume() { super.onResume(); mapView.onResume() }
    override fun onPause() { super.onPause(); mapView.onPause() }
}