package com.basicauth

import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

interface ApiService {
    @POST("register")
    suspend fun register(@Body request: RegisterRequest): Response<Unit>

    @POST("login")
    suspend fun login(@Body request: LoginRequest): Response<LoginResponse>

    @POST("forgot-password")
    suspend fun forgotPassword(@Body request: ForgotPasswordRequest): Response<Unit>

    @GET("users")
    suspend fun getUsers(@Header("Authorization") token: String): Response<List<User>>
}

data class RegisterRequest(
    val username: String,
    val email: String,
    val password: String,
    val securityQuestion: String,
    val securityAnswer: String
)

data class LoginRequest(
    val email: String,
    val password: String
)

data class LoginResponse(
    val token: String
)

data class ForgotPasswordRequest(
    val email: String,
    val securityAnswer: String,
    val newPassword: String
)

data class User(
    val id: Int,
    val username: String,
    val email: String,
    val createdAt: String
) {
    fun getFormattedCreatedAt(): String {
        val formatter = DateTimeFormatter.ofPattern("MMM dd, yyyy hh:mm a")
        val zonedDateTime = ZonedDateTime.parse(createdAt) // Parse ISO 8601 format
        val localDateTime = zonedDateTime.withZoneSameInstant(ZoneId.systemDefault()) // Convert to local timezone
        return formatter.format(localDateTime) // Format as per pattern
    }
}