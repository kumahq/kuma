import{u as h}from"./production-060535a4.js";import{L as l,_ as m}from"./kongponents.es-79677c68.js";import{u as y}from"./store-3df31b4b.js";import{d as f,c as b,o,h as a,e as c,a4 as g,u as n,w as i,f as e,g as t,t as k,b as p,p as w,k as S}from"./runtime-dom.esm-bundler-062436f2.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(w("data-v-6df60a5a"),d=d(),S(),d),K={class:"wizard-switcher"},U={class:"capitalize"},E={key:0},z={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},L={key:0},D=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),T={class:"text-center"},j={key:1},q=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},F=f({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=h(),v=y(),r=b(()=>v.getters["config/getEnvironment"]);return(G,H)=>(o(),a("div",K,[c(n(m),{ref:"emptyState","cta-is-hidden":"","is-error":!n(r),class:"my-6"},g({body:i(()=>[n(r)==="kubernetes"?(o(),a("div",E,[n(_).name===s.kubernetes?(o(),a("div",z,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):n(r)==="universal"?(o(),a("div",R,[n(_).name===s.kubernetes?(o(),a("div",L,[D,e(),t("p",T,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",j,[q,e(),t("p",A,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):p("",!0)]),_:2},[n(r)==="kubernetes"||n(r)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",U,k(n(r)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const X=x(F,[["__scopeId","data-v-6df60a5a"]]);export{X as E};
