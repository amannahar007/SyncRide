package com.syncride.app

import android.os.Bundle
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class AllServicesActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_all_services)

        findViewById<TextView>(R.id.backBtn).setOnClickListener { finish() }

        // Setup click listeners for all services
        setupServiceClickListener(R.id.serviceCab, "Shared Cab Pooling")
        setupServiceClickListener(R.id.serviceAuto, "Shared Auto Network")
        setupServiceClickListener(R.id.serviceBus, "Local Bus Transit Pass")
        setupServiceClickListener(R.id.serviceTraveller, "Intercity Traveller Sync")
        setupServiceClickListener(R.id.serviceOutstation, "Outstation Peer Pool")
        setupServiceClickListener(R.id.serviceParcel, "Shared Parcel Drop")
        setupServiceClickListener(R.id.serviceGrocery, "Grocery Sync Order")
        setupServiceClickListener(R.id.serviceStudent, "Student Campus Shuttle")
        setupServiceClickListener(R.id.serviceCorporate, "Corporate Pool Pass")
    }

    private fun setupServiceClickListener(viewId: Int, serviceName: String) {
        findViewById<TextView>(viewId)?.setOnClickListener {
            Toast.makeText(this, "Opening: $serviceName", Toast.LENGTH_SHORT).show()
        }
    }
}