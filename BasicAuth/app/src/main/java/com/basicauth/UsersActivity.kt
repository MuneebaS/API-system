package com.basicauth

import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import com.basicauth.databinding.ActivityUsersBinding
import kotlinx.coroutines.launch
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

class UsersActivity : AppCompatActivity() {
    private lateinit var binding: ActivityUsersBinding
    private lateinit var apiService: ApiService
    private lateinit var adapter: UsersAdapter

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityUsersBinding.inflate(layoutInflater)
        setContentView(binding.root)

        apiService = Retrofit.Builder()
            .baseUrl("http://10.0.2.2:8080/")
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)

        setupRecyclerView()
        loadUsers()
    }

    private fun setupRecyclerView() {
        adapter = UsersAdapter()
        binding.recyclerViewUsers.adapter = adapter
        binding.recyclerViewUsers.layoutManager = LinearLayoutManager(this)
    }

    private fun loadUsers() {
        val token = getSharedPreferences("auth", MODE_PRIVATE).getString("token", "") ?: ""

        lifecycleScope.launch {
            try {
                val response = apiService.getUsers("Bearer $token")
                if (response.isSuccessful) {
                    response.body()?.let { users ->
                        adapter.submitList(users)
                    }
                } else {
                    Toast.makeText(this@UsersActivity, "Failed to load users", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                Toast.makeText(this@UsersActivity, "Error: ${e.message}", Toast.LENGTH_SHORT).show()
            }
        }
    }
}