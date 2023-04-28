import{M as l,x as h}from"./kongponents.es-5adaddec.js";import{d as m,u as b,c as y,o,j as r,g as c,H as f,b as n,w as i,h as e,i as t,t as g,f as p,p as k,m as w}from"./index-a24b4f04.js";import{u as S}from"./store-07fabdaf.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),K={class:"wizard-switcher"},U={class:"capitalize"},E={key:0},z={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},M={key:0},j=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},H={key:1},T=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),q={class:"text-center"},A=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=b(),v=S(),a=y(()=>v.getters["config/getEnvironment"]);return(F,G)=>(o(),r("div",K,[c(n(h),{ref:"emptyState","cta-is-hidden":"","is-error":!n(a),class:"my-6"},f({body:i(()=>[n(a)==="kubernetes"?(o(),r("div",E,[n(_).name===s.kubernetes?(o(),r("div",z,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):n(a)==="universal"?(o(),r("div",R,[n(_).name===s.kubernetes?(o(),r("div",M,[j,e(),t("p",D,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",H,[T,e(),t("p",q,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):p("",!0)]),_:2},[n(a)==="kubernetes"||n(a)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",U,g(n(a)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const Q=x(A,[["__scopeId","data-v-f74b1174"]]);export{Q as E};
