import{T as l,F as h}from"./kongponents.es-76ff1c1d.js";import{d as m,u as b,c as y,o,j as r,g as c,E as f,b as n,w as i,h as e,i as t,t as g,f as p,p as k,m as w}from"./index-e1c5e7d3.js";import{u as S}from"./store-8a8250b5.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),E={class:"wizard-switcher"},K={class:"capitalize"},U={key:0},z={key:0},I=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},T={key:0},F=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),j={class:"text-center"},D={key:1},q=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},G=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=b(),v=S(),a=y(()=>v.getters["config/getEnvironment"]);return(H,J)=>(o(),r("div",E,[c(n(h),{ref:"emptyState","cta-is-hidden":"","is-error":!n(a),class:"my-6"},f({body:i(()=>[n(a)==="kubernetes"?(o(),r("div",U,[n(_).name===s.kubernetes?(o(),r("div",z,[I,e(),t("p",N,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):n(a)==="universal"?(o(),r("div",R,[n(_).name===s.kubernetes?(o(),r("div",T,[F,e(),t("p",j,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),r("div",D,[q,e(),t("p",A,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):p("",!0)])):p("",!0)]),_:2},[n(a)==="kubernetes"||n(a)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",K,g(n(a)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const Q=x(G,[["__scopeId","data-v-f74b1174"]]);export{Q as E};
