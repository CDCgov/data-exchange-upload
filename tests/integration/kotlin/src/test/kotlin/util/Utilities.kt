package util
import org.json.JSONObject
class Utilities{
    fun extractMetadataFromResponseBody(jsonResponse: String): HashMap<String, String> {
        val metadataMap = HashMap<String, String>()
        val jsonObject = JSONObject(jsonResponse)
        val reportsArray = jsonObject.getJSONArray("reports")

        for (i in 0 until reportsArray.length()) {
            val report = reportsArray.getJSONObject(i)
            val content = report.getJSONObject("content")

            if (content.has("metadata")) {
                val metadata = content.getJSONObject("metadata")
                val keysIterator = metadata.keys()
                while (keysIterator.hasNext()) {
                    val key = keysIterator.next() as String
                    val value = metadata.getString(key)
                    metadataMap[key] = value
                }
            }
        }

        return metadataMap
    }

}