import{u as h}from"./vue-router-d8e03a07.js";import{P as l,$ as m}from"./kongponents.es-c2485d1e.js";import{u as y}from"./store-96085224.js";import{d as f,c as b,o,h as a,e as c,a4 as g,u as n,w as i,f as e,g as t,t as k,b as p,p as w,j as S}from"./runtime-dom.esm-bundler-32659b48.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(w("data-v-6df60a5a"),d=d(),S(),d),K={class:"wizard-switcher"},U={class:"capitalize"},E={key:0},z={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},P={key:0},$=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),j={class:"text-center"},D={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),q={class:"text-center"},A=f({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=h(),v=y(),r=b(()=>v.getters["config/getEnvironment"]);return(F,G)=>(o(),a("div",K,[c(n(m),{ref:"emptyState","cta-is-hidden":"","is-error":!n(r),class:"my-6"},g({body:i(()=>[n(r)==="kubernetes"?(o(),a("div",E,[n(_).name===s.kubernetes?(o(),a("div",z,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):n(r)==="universal"?(o(),a("div",R,[n(_).name===s.kubernetes?(o(),a("div",P,[$,e(),t("p",j,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",D,[T,e(),t("p",q,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):p("",!0)]),_:2},[n(r)==="kubernetes"||n(r)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",U,k(n(r)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const Q=x(A,[["__scopeId","data-v-6df60a5a"]]);export{Q as E};
