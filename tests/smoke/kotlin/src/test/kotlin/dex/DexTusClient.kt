package dex

import io.tus.java.client.TusClient
import java.net.HttpURLConnection

class DexTusClient : TusClient() {
    lateinit var connection: HttpURLConnection

    override fun prepareConnection(connection: HttpURLConnection) {
        super.prepareConnection(connection)
        this.connection = connection
    }
}