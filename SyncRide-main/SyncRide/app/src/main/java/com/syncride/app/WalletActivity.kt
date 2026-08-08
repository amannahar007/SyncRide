package com.syncride.app

import android.content.Context
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity

class WalletActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_wallet)

        val sharedPref = getSharedPreferences("SyncRidePrefs", Context.MODE_PRIVATE)
        val balance = sharedPref.getFloat("WALLET_BALANCE", 0f)
        findViewById<TextView>(R.id.walletBalance).text = "₹${"%.2f".format(balance)}"

        findViewById<TextView>(R.id.backButton).setOnClickListener { finish() }
        findViewById<Button>(R.id.btnAddMoney).setOnClickListener {
            Toast.makeText(this, "Opening Payment Gateway...", Toast.LENGTH_SHORT).show()
        }
    }
}