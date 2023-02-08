import{P as c,$ as i}from"./kongponents.es-c2485d1e.js";import{u as m}from"./store-15db9444.js";import{u as h}from"./index-ad2ab4aa.js";import{d as M,c as U,o as A,h as f,e as o,w as n,f as s,g as t,u as e,t as r,p as R,j as v}from"./runtime-dom.esm-bundler-32659b48.js";import{_ as C}from"./_plugin-vue_export-helper-c27b6911.js";const l=_=>(R("data-v-189f4a92"),_=_(),v(),_),b={class:"resource-list"},y=l(()=>t("p",null,[s(`
          We can create multiple isolated Mesh resources (i.e. per application/`),t("wbr"),s("team/"),t("wbr"),s(`business unit).
        `)],-1)),E={class:"resource-list-actions mt-4"},K=l(()=>t("p",null,`
          We need a data plane proxy for each replica of our services within a Mesh resource.
        `,-1)),g={class:"resource-list-actions mt-4"},P={class:"resource-list-actions mt-4"},T=["href"],S=["href"],k=["href"],x=M({__name:"MeshResources",setup(_){const a=h(),p=m(),u=U(()=>({name:p.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"}));return(d,D)=>(A(),f("div",b,[o(e(i),{title:"Create a virtual mesh"},{body:n(()=>[y,s(),t("div",E,[o(e(c),{icon:"plus",appearance:"creation",to:{name:"create-mesh"}},{default:n(()=>[s(`
            Create mesh
          `)]),_:1})])]),_:1}),s(),o(e(i),{title:"Connect data plane proxies"},{body:n(()=>[K,s(),t("div",g,[o(e(c),{to:e(u),appearance:"primary"},{default:n(()=>[s(`
            Get started
          `)]),_:1},8,["to"])])]),_:1}),s(),o(e(i),{title:`Apply ${e(a)("KUMA_PRODUCT_NAME")} policies`},{body:n(()=>[t("p",null,`
          We can apply `+r(e(a)("KUMA_PRODUCT_NAME"))+` policies to secure, observe, route and manage the Mesh and its data plane proxies.
        `,1),s(),t("div",P,[o(e(c),{to:`${e(a)("KUMA_DOCS_URL")}/policies/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,appearance:"primary",target:"_blank"},{default:n(()=>[s(`
            Explore policies
          `)]),_:1},8,["to"])])]),_:1},8,["title"]),s(),o(e(i),{title:"Resources"},{body:n(()=>[t("p",null,`
          Join the `+r(e(a)("KUMA_PRODUCT_NAME"))+` community and ask questions:
        `,1),s(),t("ul",null,[t("li",null,[t("a",{href:`${e(a)("KUMA_DOCS_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+` Documentation
            `,9,T)]),s(),t("li",null,[t("a",{href:`${e(a)("KUMA_CHAT_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+`  Community Chat
            `,9,S)]),s(),t("li",null,[t("a",{href:`https://github.com/kumahq/kuma?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+` GitHub Repository
            `,9,k)])])]),_:1})]))}});const Q=C(x,[["__scopeId","data-v-189f4a92"]]);export{Q as M};
