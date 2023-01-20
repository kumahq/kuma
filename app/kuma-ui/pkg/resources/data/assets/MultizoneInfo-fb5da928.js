import{bU as i,k as m,ck as l,P as u,$ as _,m as d,cc as t,o as K,i as $,w as n,a,b as e,e as s,bW as M}from"./index-2d2ce866.js";const f=i(),g={name:"MultizoneInfo",env:f,productName:l,components:{KButton:u,KEmptyState:_,KIcon:d}},y=s("p",null,[e(`
        To access this page, you must be running in `),s("strong",null,"Multi-Zone"),e(` mode.
      `)],-1);function S(o,U,b,B,E,N){const c=t("KIcon"),r=t("KButton"),p=t("KEmptyState");return K(),$(p,null,{title:n(()=>[a(c,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),e(),s("p",null,M(o.$options.productName)+" is running in Standalone mode.",1)]),message:n(()=>[y]),cta:n(()=>[a(r,{to:`${o.$options.env("KUMA_DOCS_URL")}/documentation/deployments/?${o.$options.env("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank",appearance:"primary"},{default:n(()=>[e(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const v=m(g,[["render",S]]);export{v as M};
