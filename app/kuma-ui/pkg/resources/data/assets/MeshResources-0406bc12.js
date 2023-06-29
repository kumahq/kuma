import{S as i,U as c}from"./kongponents.es-a99534bb.js";import{l as h,f as m,h as M}from"./RouteView.vue_vue_type_script_setup_true_lang-4999f19d.js";import{d as U,c as A,o as R,e as f,h as o,w as n,g as s,k as t,b as e,t as r,p as b,m as v}from"./index-0b8ed13f.js";const l=_=>(b("data-v-88c956eb"),_=_(),v(),_),C={class:"kcard-switcher"},y=l(()=>t("p",null,[s(`
          We can create multiple isolated Mesh resources (i.e. per application/`),t("wbr"),s("team/"),t("wbr"),s(`business unit).
        `)],-1)),E={class:"resource-list-actions mt-4"},K=l(()=>t("p",null,`
          We need a data plane proxy for each replica of our services within a Mesh resource.
        `,-1)),S={class:"resource-list-actions mt-4"},g={class:"resource-list-actions mt-4"},k=["href"],T=["href"],x=["href"],D=U({__name:"MeshResources",setup(_){const a=h(),p=m(),u=A(()=>({name:p.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"}));return(d,P)=>(R(),f("div",C,[o(e(c),{title:"Create a virtual mesh"},{body:n(()=>[y,s(),t("div",E,[o(e(i),{icon:"plus",appearance:"creation",to:{name:"create-mesh"}},{default:n(()=>[s(`
            Create mesh
          `)]),_:1})])]),_:1}),s(),o(e(c),{title:"Connect data plane proxies"},{body:n(()=>[K,s(),t("div",S,[o(e(i),{to:u.value,appearance:"primary"},{default:n(()=>[s(`
            Get started
          `)]),_:1},8,["to"])])]),_:1}),s(),o(e(c),{title:`Apply ${e(a)("KUMA_PRODUCT_NAME")} policies`},{body:n(()=>[t("p",null,`
          We can apply `+r(e(a)("KUMA_PRODUCT_NAME"))+` policies to secure, observe, route and manage the Mesh and its data plane proxies.
        `,1),s(),t("div",g,[o(e(i),{to:`${e(a)("KUMA_DOCS_URL")}/policies/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,appearance:"primary",target:"_blank"},{default:n(()=>[s(`
            Explore policies
          `)]),_:1},8,["to"])])]),_:1},8,["title"]),s(),o(e(c),{title:"Resources"},{body:n(()=>[t("p",null,`
          Join the `+r(e(a)("KUMA_PRODUCT_NAME"))+` community and ask questions:
        `,1),s(),t("ul",null,[t("li",null,[t("a",{href:`${e(a)("KUMA_DOCS_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+` Documentation
            `,9,k)]),s(),t("li",null,[t("a",{href:`${e(a)("KUMA_CHAT_URL")}/?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+`  Community Chat
            `,9,T)]),s(),t("li",null,[t("a",{href:`https://github.com/kumahq/kuma?${e(a)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},r(e(a)("KUMA_PRODUCT_NAME"))+` GitHub Repository
            `,9,x)])])]),_:1})]))}});const w=M(D,[["__scopeId","data-v-88c956eb"]]);export{w as M};
