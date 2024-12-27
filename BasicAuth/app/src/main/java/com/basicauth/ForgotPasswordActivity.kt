package com.basicauth

import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.basicauth.databinding.ActivityForgotPasswordBinding
import kotlinx.coroutines.launch
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

class ForgotPasswordActivity : AppCompatActivity() {
    private lateinit var binding: ActivityForgotPasswordBinding
    private lateinit var apiService: ApiService

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityForgotPasswordBinding.inflate(layoutInflater)
        setContentView(binding.root)

        apiService = Retrofit.Builder()
            .baseUrl("http://10.0.2.2:8080/")
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)

            binding.buttonResetPassword.setOnClickListener {
                val email = binding.editTextEmail.text.toString()
                val securityAnswer = binding.editTextSecurityAnswer.text.toString()
                val newPassword = binding.editTextNewPassword.text.toString()

                if (email.isNotEmpty() && securityAnswer.isNotEmpty() && newPassword.isNotEmpty()) {
                    resetPassword(email, securityAnswer, newPassword)
                } else {
                    Toast.makeText(this, "Please fill all fields", Toast.LENGTH_SHORT).show()
                }
            }
    }

    private fun resetPassword(email: String, securityAnswer: String, newPassword: String) {
        lifecycleScope.launch {
            try {
                val request = ForgotPasswordRequest(email, securityAnswer, newPassword)
                val response = apiService.forgotPassword(request)

                if (response.isSuccessful) {
                    Toast.makeText(this@ForgotPasswordActivity,
                        "Password reset successful", Toast.LENGTH_SHORT).show()
                    finish()
                } else {
                    Toast.makeText(this@ForgotPasswordActivity,
                        "Failed to reset password", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                Toast.makeText(this@ForgotPasswordActivity,
                    "Error: ${e.message}", Toast.LENGTH_SHORT).show()
            }
        }
    }
}