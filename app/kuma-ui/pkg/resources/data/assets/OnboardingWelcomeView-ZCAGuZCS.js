import{d as v,a as c,o as r,b as d,w as n,e as o,f as e,t as l,m as t,F as b,G as y,q as u,K as V,p as N,c as m,H as x,x as I,y as O,_ as C}from"./index-KOnKkPpw.js";import{O as k,a as R,b as A}from"./OnboardingPage-I9JWyT9d.js";const p=s=>(I("data-v-6d3e602f"),s=s(),O(),s),S=p(()=>t("strong",null,"few minutes",-1)),z={"data-testid":"kuma-environment"},B=p(()=>t("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),E={class:"item-status-list-wrapper"},L={class:"item-status-list"},T={class:"circle mr-2"},W=v({__name:"OnboardingWelcomeView",setup(s){return(K,F)=>{const _=c("RouteTitle"),g=c("AppView"),f=c("RouteView");return r(),d(f,{name:"onboarding-welcome-view"},{default:n(({env:h,t:a,can:w})=>[o(_,{title:a("onboarding.routes.welcome.title",{name:a("common.product.name")}),render:!1},null,8,["title"]),e(),o(g,null,{default:n(()=>[t("div",null,[o(k,null,{header:n(()=>[o(R,null,{title:n(()=>[e(`
                Welcome to `+l(a("common.product.name")),1)]),description:n(()=>[t("p",null,[e(`
                  Congratulations on downloading `+l(a("common.product.name"))+"! You are just a ",1),S,e(` away from getting your service mesh fully online.
                `)]),e(),t("p",null,[e(`
                  We have automatically detected that you are running on `),t("strong",z,l(a(`common.product.environment.${h("KUMA_ENVIRONMENT")}`)),1),e(`.
                `)])]),_:2},1024)]),content:n(()=>[B,e(),t("div",E,[t("ul",L,[(r(!0),m(b,null,x([{name:`Run ${a("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...w("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],i=>(r(),m("li",{key:i.name},[t("span",T,[i.status?(r(),d(u(y),{key:0,size:u(V)},null,8,["size"])):N("",!0)]),e(" "+l(i.name),1)]))),128))])])]),navigation:n(()=>[o(A,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}}),M=C(W,[["__scopeId","data-v-6d3e602f"]]);export{M as default};
