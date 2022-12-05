/*
When mounting a component via vue-test-utils, it is unaware of (and concerned by)
The absence of NuxtLinks (and other Nuxt-Primitive components that aren't) 
available to the vue instance rendering our component. To date no libraries exist
to test Nuxt components nicely, and it's led to heavy setup reliance [1][2][3].

Additionally, the vue-test-utils framework does allow you to use a custom set of
options to run with vue, but doesn't actually respect the one that you'd need to
set to mark the Nuxt specific components as custom/framework components [4].

Using the s(t)ubstitution approach in [5] allows us to work through most of these
issues, since there are Vue analogues for most of nuxt's custom components. We
configure those globally in this file, rather than in each test. However, that 
doesn't stop the warnings we receive, and since vue can't be made to suppress
these warnings [6] we suppress them within jest instead (see silence-some-warnings.ts). 

If [5] is ever fixed, we can remove this by setting vue-compilation options within
the mount/shallowMount method in our tests:

global: {
    config: {
        compilerOptions: {
        isCustomElement: isNuxtComponent,            
        }
    },
},

[1] https://stackoverflow.com/questions/65165605
[2] https://blog.logrocket.com/component-testing-in-nuxt-js
[3] https://dev.to/alousilva/how-to-mock-nuxt-client-only-component-with-jest-47da
[4] https://github.com/vuejs/vue-test-utils/issues/1865
[5] https://stackoverflow.com/questions/49665571
[6] https://stackoverflow.com/questions/43933405
*/
import { config, RouterLinkStub } from '@vue/test-utils'

config.global.stubs = {
    ...config.global.stubs,
    'NuxtLink': RouterLinkStub 
}
