import{d as p,m,i as y,j as b,o,c as a,a as c,Q as g,u as n,w as i,e,g as t,y as f,E as l,x as v,M as k,N as w,O as S,_ as x}from"./index-521c0f24.js";const u=d=>(w("data-v-f74b1174"),d=d(),S(),d),E={class:"wizard-switcher"},K={class:"capitalize"},U={key:0},z={key:0},N=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              and we are going to be showing you instructions for Kubernetes unless you
              decide to visualize the instructions for Universal.
            `)],-1)),I={class:"text-center"},V=u(()=>t("br",null,null,-1)),W={key:1},B=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Kubernetes environment"),e(`,
              but you are viewing instructions for Universal.
            `)],-1)),C={class:"text-center"},R={key:1},M={key:0},j=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              but you are viewing instructions for Kubernetes.
            `)],-1)),D={class:"text-center"},O={key:1},Q=u(()=>t("p",null,[e(`
              We have detected that you are running on a `),t("strong",null,"Universal environment"),e(`,
              and we are going to be showing you instructions for Universal unless you
              decide to visualize the instructions for Kubernetes.
            `)],-1)),T={class:"text-center"},q=p({__name:"EnvironmentSwitcher",setup(d){const s={kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"},_=m(),h=y(),r=b(()=>h.getters["config/getEnvironment"]);return(A,F)=>(o(),a("div",E,[c(n(k),{ref:"emptyState","cta-is-hidden":"","is-error":!n(r),class:"my-6"},g({body:i(()=>[n(r)==="kubernetes"?(o(),a("div",U,[n(_).name===s.kubernetes?(o(),a("div",z,[N,e(),t("p",I,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch to`),V,e(`
                Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",W,[B,e(),t("p",C,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):n(r)==="universal"?(o(),a("div",R,[n(_).name===s.kubernetes?(o(),a("div",M,[j,e(),t("p",D,[c(n(l),{to:{name:s.universal},appearance:"secondary"},{default:i(()=>[e(`
                Switch back to Universal instructions
              `)]),_:1},8,["to"])])])):n(_).name===s.universal?(o(),a("div",O,[Q,e(),t("p",T,[c(n(l),{to:{name:s.kubernetes},appearance:"secondary"},{default:i(()=>[e(`
                Switch to
                Kubernetes instructions
              `)]),_:1},8,["to"])])])):v("",!0)])):v("",!0)]),_:2},[n(r)==="kubernetes"||n(r)==="universal"?{name:"title",fn:i(()=>[e(`
        Running on `),t("span",K,f(n(r)),1)]),key:"0"}:void 0]),1032,["is-error"])]))}});const H=x(q,[["__scopeId","data-v-f74b1174"]]);export{H as E};
