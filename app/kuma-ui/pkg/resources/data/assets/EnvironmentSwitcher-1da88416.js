import{a as l,M as h}from"./kongponents.es-3ba46133.js";import{u as m}from"./production-4c848fca.js";import{u as b}from"./store-e9d4becb.js";import{d as y,c as f,o,h as r,e as c,a4 as g,u as n,w as i,f as e,g as t,t as k,b as p,p as w,k as S}from"./runtime-dom.esm-bundler-9284044f.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(w("data-v-f74b1174"),d=d(),S(),d),E={class:"wizard-switcher"},K={class:"capitalize"},U={key:0},z={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},M={key:0},D=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),T={class:"text-center"},j={key:1},q=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},F=y({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=m(),v=b(),a=f(()=>v.getters["config/getEnvironment"]);return(G,H)=>(o(),r("div",E,[c(n(h),{ref:"emptyState","cta-is-hidden":"","is-error":!n(a),class:"my-6"},g({body:i(()=>[n(a)==="kubernetes"?(o(),r("div",U,[n(_).name===s.kubernetes?(o(),r("div",z,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):n(a)==="universal"?(o(),r("div",R,[n(_).name===s.kubernetes?(o(),r("div",M,[D,e(),t("p",T,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",j,[q,e(),t("p",A,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):p("",!0)]),_:2},[n(a)==="kubernetes"||n(a)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",K,k(n(a)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const X=x(F,[["__scopeId","data-v-f74b1174"]]);export{X as E};
