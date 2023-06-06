import{x as _,Z as h}from"./kongponents.es-33458096.js";import{d as m,u as y,c as b,o,j as a,g as c,y as f,w as i,h as e,i as n,t as g,e as t,f as v,p as k,m as w}from"./index-aa84cb87.js";import{u as S}from"./store-3ba1feee.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),K={class:"wizard-switcher"},U={class:"capitalize"},E={key:0},z={key:0},I=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>n("br",null,null,-1)),W={key:1},B=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},Z={key:0},j=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},T={key:1},q=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},F=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},l=y(),p=S(),r=b(()=>p.getters["config/getEnvironment"]);return(G,H)=>(o(),a("div",K,[c(t(h),{ref:"emptyState","cta-is-hidden":"","is-error":!r.value,class:"my-6"},f({body:i(()=>[r.value==="kubernetes"?(o(),a("div",E,[t(l).name===s.kubernetes?(o(),a("div",z,[I,e(),n("p",N,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",W,[B,e(),n("p",C,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):r.value==="universal"?(o(),a("div",R,[t(l).name===s.kubernetes?(o(),a("div",Z,[j,e(),n("p",D,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",T,[q,e(),n("p",A,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[r.value==="kubernetes"||r.value==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),n("span",U,g(r.value),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const P=x(F,[["__scopeId","data-v-f74b1174"]]);export{P as E};
