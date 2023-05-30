import{_,z as h}from"./kongponents.es-c0dfafb1.js";import{d as m,u as y,c as b,o,j as a,g as c,z as f,w as i,h as e,i as n,t as g,e as t,f as v,p as k,m as w}from"./index-cf8f17f6.js";import{u as S}from"./store-a56707da.js";import{_ as x}from"./_plugin-vue_export-helper-c27b6911.js";const u=d=>(k("data-v-f74b1174"),d=d(),w(),d),z={class:"wizard-switcher"},K={class:"capitalize"},U={key:0},E={key:0},I=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),N={class:"text-center"},V=u(()=>n("br",null,null,-1)),W={key:1},B=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},j={key:0},D=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),T={class:"text-center"},q={key:1},A=u(()=>n("p",null,[e(`
              We have detected that you are running on a `),n("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),F={class:"text-center"},G=m({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},l=y(),p=S(),r=b(()=>p.getters["config/getEnvironment"]);return(H,J)=>(o(),a("div",z,[c(t(h),{ref:"emptyState","cta-is-hidden":"","is-error":!r.value,class:"my-6"},f({body:i(()=>[r.value==="kubernetes"?(o(),a("div",U,[t(l).name===s.kubernetes?(o(),a("div",E,[I,e(),n("p",N,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",W,[B,e(),n("p",C,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):r.value==="universal"?(o(),a("div",R,[t(l).name===s.kubernetes?(o(),a("div",j,[D,e(),n("p",T,[c(t(_),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):t(l).name===s.universal?(o(),a("div",q,[A,e(),n("p",F,[c(t(_),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[r.value==="kubernetes"||r.value==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),n("span",K,g(r.value),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const Q=x(G,[["__scopeId","data-v-f74b1174"]]);export{Q as E};
