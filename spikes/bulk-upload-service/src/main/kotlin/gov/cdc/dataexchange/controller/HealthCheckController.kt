package gov.cdc.dataexchange.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController
import java.util.*

@RestController
class HelloController {
    @GetMapping("/")
    fun index(): String {
        return "Hello There. You pinged me at ${Date()}"
    }
}