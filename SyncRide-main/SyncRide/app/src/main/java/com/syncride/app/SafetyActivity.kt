package com.syncride.app

import android.os.Bundle
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.cardview.widget.CardView

class SafetyActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_safety)

        findViewById<TextView>(R.id.backButton).setOnClickListener { finish() }

        findViewById<CardView>(R.id.btnSOS).setOnClickListener {
            Toast.makeText(this, "SOS Alert Sent to Emergency Services!", Toast.LENGTH_LONG).show()
        }
        findViewById<LinearLayout>(R.id.btnShareLocation).setOnClickListener {
            Toast.makeText(this, "Live Location Link copied to clipboard", Toast.LENGTH_SHORT).show()
        }
        findViewById<LinearLayout>(R.id.btnEmergencyContacts).setOnClickListener {
            Toast.makeText(this, "Opening Emergency Contacts", Toast.LENGTH_SHORT).show()
        }
    }
}