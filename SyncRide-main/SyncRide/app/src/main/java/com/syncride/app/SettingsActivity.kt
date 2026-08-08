package com.syncride.app

import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class SettingsActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_settings)

        findViewById<TextView>(R.id.backButton).setOnClickListener { finish() }

        findViewById<TextView>(R.id.settingProfile).setOnClickListener {
            Toast.makeText(this, "Profile Details", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.settingAddress).setOnClickListener {
            Toast.makeText(this, "Manage Addresses", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.settingNotifications).setOnClickListener {
            Toast.makeText(this, "Notification Settings", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.settingPrivacy).setOnClickListener {
            Toast.makeText(this, "Privacy & Safety", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.settingHelp).setOnClickListener {
            Toast.makeText(this, "Opening Help Center", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.settingAbout).setOnClickListener {
            Toast.makeText(this, "SyncRide v1.0.2", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.btnLogout).setOnClickListener {
            val sharedPref = getSharedPreferences("SyncRidePrefs", Context.MODE_PRIVATE)
            sharedPref.edit().clear().apply()
            
            val intent = Intent(this, LoginActivity::class.java)
            intent.flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
            startActivity(intent)
            finish()
        }
    }
}