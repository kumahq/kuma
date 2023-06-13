import{T as _,Z as p}from"./kongponents.es-b630df1e.js";import{d as m,u as b,c as y,o,f as a,i as c,m as f,b as i,h as e,l as n,t as g,e as t,g as v,p as k,n as w}from"./index-64044ff8.js";import{f as S,h as x}from"./RouteView.vue_vue_type_script_setup_true_lang-75502ce3.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),K={class:"wizard-switcher"},U={class:"capitalize"},E={key:0},z={key:0},I=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>n("br",null,null,-1)),W={key:1},B=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},T={key:0},Z=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},j={key:1},q=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),A={class:"text-center"},F=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},l=b(),h=S(),r=y(()=>h.getters["config/getEnvironment"]);return(G,H)=>(o(),a("div",K,[c(t(p),{ref:"emptyState","cta-is-hidden":"","is-error":!r.value,class:"my-6"},f({body:i(()=>[r.value==="kubernetes"?(o(),a("div",E,[t(l).name===s.kubernetes?(o(),a("div",z,[I,e(),n("p",N,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",W,[B,e(),n("p",C,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):r.value==="universal"?(o(),a("div",R,[t(l).name===s.kubernetes?(o(),a("div",T,[Z,e(),n("p",D,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",j,[q,e(),n("p",A,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[r.value==="kubernetes"||r.value==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),n("span",U,g(r.value),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const O=x(F,[["__scopeId","data-v-f74b1174"]]);export{O as E};
