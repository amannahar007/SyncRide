package com.syncride.app

import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.cardview.widget.CardView

class ProfileActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_profile)

        val sharedPref = getSharedPreferences("SyncRidePrefs", Context.MODE_PRIVATE)
        val userName = sharedPref.getString("USER_NAME", "User")
        val walletBalance = sharedPref.getFloat("WALLET_BALANCE", 0f)
        val ridesCount = sharedPref.getInt("RIDES_COUNT", 128)

        findViewById<TextView>(R.id.profileName).text = userName
        findViewById<TextView>(R.id.profileWalletBalance).text = "₹${"%.2f".format(walletBalance)}"
        findViewById<TextView>(R.id.profileRidesCount).text = ridesCount.toString()

        findViewById<TextView>(R.id.backButton).setOnClickListener { finish() }

        findViewById<TextView>(R.id.menuMyRides).setOnClickListener {
            val intent = Intent(this, RideHistoryActivity::class.java)
            startActivity(intent)
        }
        findViewById<TextView>(R.id.menuPayment).setOnClickListener {
            val intent = Intent(this, WalletActivity::class.java)
            startActivity(intent)
        }
        findViewById<TextView>(R.id.menuSafety).setOnClickListener {
            val intent = Intent(this, SafetyActivity::class.java)
            startActivity(intent)
        }
        findViewById<CardView>(R.id.menuRefer).setOnClickListener {
            Toast.makeText(this, "Generating Referral Code...", Toast.LENGTH_SHORT).show()
        }
        findViewById<TextView>(R.id.menuSettings).setOnClickListener {
            val intent = Intent(this, SettingsActivity::class.java)
            startActivity(intent)
        }
    }
}