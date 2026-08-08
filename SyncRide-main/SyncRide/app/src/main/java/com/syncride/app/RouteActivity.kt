package com.syncride.app

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.text.Editable
import android.text.TextWatcher
import android.view.LayoutInflater
import android.view.View
import android.view.inputmethod.InputMethodManager
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class RouteActivity : AppCompatActivity() {

    private lateinit var pickupInput: EditText
    private lateinit var destinationInput: EditText
    private lateinit var btnClearInput: TextView
    private lateinit var resultsContainer: LinearLayout
    private lateinit var savedPlacesSection: View

    private var activeField: EditText? = null

    private val mockPlaces = listOf(
        "Geetanjali Institute of Technical Studies (GITS)",
        "Udiapole Circle, Udaipur",
        "Dabok Airport, Rajasthan",
        "Fateh Sagar Lake, Udaipur",
        "City Palace, Udaipur",
        "Sukhadia Circle",
        "Udaipur City Railway Station",
        "Lake Pichola",
        "Sajjangarh Monsoon Palace",
        "Jagdish Temple",
        "Bapu Bazaar",
        "Shastri Circle",
        "Pacific University",
        "MLSU University Campus",
        "Paras Cinema",
        "Celebration Mall",
        "Udaipur Bus Stand",
        "Hiran Magri Sector 4",
        "Fatehpura Circle"
    )

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_route)

        pickupInput = findViewById(R.id.pickupInput)
        destinationInput = findViewById(R.id.destinationInput)
        btnClearInput = findViewById(R.id.btnClearInput)
        resultsContainer = findViewById(R.id.resultsContainer)
        savedPlacesSection = findViewById(R.id.savedPlacesSection)

        activeField = destinationInput
        destinationInput.requestFocus()
        showKeyboard(destinationInput)

        findViewById<TextView>(R.id.backBtn).setOnClickListener { finish() }

        pickupInput.setOnFocusChangeListener { _, hasFocus ->
            if (hasFocus) {
                activeField = pickupInput
                btnClearInput.visibility = if (pickupInput.text.isNotEmpty()) View.VISIBLE else View.GONE
            }
        }

        destinationInput.setOnFocusChangeListener { _, hasFocus ->
            if (hasFocus) {
                activeField = destinationInput
                btnClearInput.visibility = if (destinationInput.text.isNotEmpty()) View.VISIBLE else View.GONE
            }
        }

        btnClearInput.setOnClickListener {
            activeField?.text?.clear()
        }

        findViewById<View>(R.id.btnSetOnMap).setOnClickListener {
            passFinalResult("MAP_PICKER")
        }

        findViewById<View>(R.id.btnHome).setOnClickListener { selectLocation("Home") }
        findViewById<View>(R.id.btnWork).setOnClickListener { selectLocation("Work") }

        setupSearchLogic()
        showInitialRecentList()
    }

    private fun setupSearchLogic() {
        val textWatcher = object : TextWatcher {
            override fun beforeTextChanged(s: CharSequence?, start: Int, count: Int, after: Int) {}
            override fun onTextChanged(s: CharSequence?, start: Int, before: Int, count: Int) {
                val query = s.toString().trim()
                if (query.isNotEmpty()) {
                    btnClearInput.visibility = View.VISIBLE
                    savedPlacesSection.visibility = View.GONE
                    filterResults(query)
                } else {
                    btnClearInput.visibility = View.GONE
                    savedPlacesSection.visibility = View.VISIBLE
                    showInitialRecentList()
                }
            }
            override fun afterTextChanged(s: Editable?) {}
        }
        
        pickupInput.addTextChangedListener(textWatcher)
        destinationInput.addTextChangedListener(textWatcher)
    }

    private fun selectLocation(name: String) {
        activeField?.setText(name)
        if (activeField == pickupInput) {
            destinationInput.requestFocus()
        } else {
            validateAndComplete()
        }
    }

    private fun validateAndComplete() {
        val pickup = pickupInput.text.toString().trim()
        val dest = destinationInput.text.toString().trim()
        
        if (pickup.isNotEmpty() && dest.isNotEmpty()) {
            passFinalResult(dest, pickup)
        }
    }

    private fun filterResults(query: String) {
        resultsContainer.removeAllViews()
        val filtered = mockPlaces.filter { it.contains(query, ignoreCase = true) }
        if (filtered.isEmpty()) {
            addNoResultsView()
        } else {
            filtered.forEach { addResultItem(it) }
        }
    }

    private fun showInitialRecentList() {
        resultsContainer.removeAllViews()
        mockPlaces.take(5).forEach { addResultItem(it, isRecent = true) }
    }

    private fun addResultItem(name: String, isRecent: Boolean = false) {
        val view = LayoutInflater.from(this).inflate(android.R.layout.simple_list_item_2, resultsContainer, false)
        val text1 = view.findViewById<TextView>(android.R.id.text1)
        val text2 = view.findViewById<TextView>(android.R.id.text2)

        text1.text = if (isRecent) "🕒 $name" else "📍 $name"
        text1.setTextColor(resources.getColor(R.color.black, theme))
        text1.textSize = 15f
        
        text2.text = if (isRecent) "Recent Destination" else "Udaipur, Rajasthan"
        text2.textSize = 12f

        view.setPadding(48, 32, 48, 32)
        view.setOnClickListener { selectLocation(name) }
        resultsContainer.addView(view)
    }

    private fun addNoResultsView() {
        val tv = TextView(this)
        tv.text = "No matching locations found"
        tv.setPadding(64, 64, 64, 64)
        tv.gravity = android.view.Gravity.CENTER
        tv.setTextColor(resources.getColor(R.color.black, theme))
        resultsContainer.addView(tv)
    }

    private fun passFinalResult(destination: String, pickup: String = "Current Location") {
        val intent = Intent().apply {
            putExtra("SELECTED_DESTINATION", destination)
            putExtra("PICKUP_LOCATION", pickup)
        }
        setResult(Activity.RESULT_OK, intent)
        finish()
    }

    private fun showKeyboard(view: View) {
        val imm = getSystemService(Context.INPUT_METHOD_SERVICE) as InputMethodManager
        imm.showSoftInput(view, InputMethodManager.SHOW_IMPLICIT)
    }
}