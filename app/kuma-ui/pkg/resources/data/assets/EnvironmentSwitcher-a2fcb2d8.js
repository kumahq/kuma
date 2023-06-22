import{S as _,U as p}from"./kongponents.es-bba90403.js";import{d as m,u as b,c as y,o,e as a,h as c,l as f,w as i,g as e,k as n,t as g,b as t,f as v,p as k,m as w}from"./index-9d631905.js";import{f as S,h as x}from"./RouteView.vue_vue_type_script_setup_true_lang-76145142.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),U={class:"wizard-switcher"},K={class:"capitalize"},E={key:0},z={key:0},I=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>n("br",null,null,-1)),W={key:1},B=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},D={key:0},T=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),j={class:"text-center"},q={key:1},A=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),F={class:"text-center"},G=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},l=b(),h=S(),r=y(()=>h.getters["config/getEnvironment"]);return(H,J)=>(o(),a("div",U,[c(t(p),{ref:"emptyState","cta-is-hidden":"","is-error":!r.value,class:"my-6"},f({body:i(()=>[r.value==="kubernetes"?(o(),a("div",E,[t(l).name===s.kubernetes?(o(),a("div",z,[I,e(),n("p",N,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",W,[B,e(),n("p",C,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):r.value==="universal"?(o(),a("div",R,[t(l).name===s.kubernetes?(o(),a("div",D,[T,e(),n("p",j,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",q,[A,e(),n("p",F,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[r.value==="kubernetes"||r.value==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),n("span",K,g(r.value),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const P=x(G,[["__scopeId","data-v-f74b1174"]]);export{P as E};
