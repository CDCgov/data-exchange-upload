package org.example

import com.github.doyaaaaaken.kotlincsv.dsl.csvReader
import org.example.model.Reupload
import java.io.File

fun main() {
    // First, read in input.csv.
    val inputCsv = getFileFromResources("input.csv")
    val reuploads = readInputCsv(inputCsv)
    println("Reuploading ${reuploads.size} file(s)")
}

fun readInputCsv(inputFile: File): List<Reupload> {
    return csvReader().readAllWithHeader(inputFile).map {
        Reupload(
            it["src"]!!,
            it["dest"]!!,
            it["srcaccountid"]!!
        )
    }
}

fun getFileFromResources(filename: String): File {
    return File(object {}::class.java.classLoader.getResource(filename)?.file!!)
}