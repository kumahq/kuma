import{d as I,I as N,F as k,o as c,m as b,w as i,c as v,G as L,E as C,e as h,a as d,b as o,k as s,t as p,H as V,J as $,l as S,K as q,p as O,L as B,M,q as R}from"./index-C4IVBmnO.js";import{O as A,a as E,b as T}from"./OnboardingPage-DYip3RDf.js";const H=["aria-hidden"],x='<path d="M9.7 18.025L4 12.325L5.425 10.9L9.7 15.175L18.875 6L20.3 7.425L9.7 18.025Z" fill="currentColor"/>',W=I({__name:"CheckIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:N,validator:e=>{if(typeof e=="number"&&e>0)return!0;if(typeof e=="string"){const n=String(e).replace(/px/gi,""),r=Number(n);if(r&&!isNaN(r)&&Number.isInteger(r)&&r>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(e){const n=e,r=k(()=>{if(typeof n.size=="number"&&n.size>0)return`${n.size}px`;if(typeof n.size=="string"){const a=String(n.size).replace(/px/gi,""),t=Number(a);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return`${t}px`}return N}),m=k(()=>({boxSizing:"border-box",color:n.color,display:n.display,flexShrink:"0",height:r.value,lineHeight:"0",width:r.value})),g=a=>{const t={},l=Math.random().toString(36).substring(2,12);return a.replace(/id="([^"]+)"/g,(y,u)=>{const w=`${l}-${u}`;return t[u]=w,`id="${w}"`}).replace(/#([^\s^")]+)/g,(y,u)=>t[u]?`#${t[u]}`:y)},f=a=>{const t={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return a.replace(/[<>"'`]/g,l=>t[l])},_=`${n.title?`<title data-testid="kui-icon-svg-title">${f(n.title)}</title>`:""}${n.staticIds?x:g(x)}`;return(a,t)=>(c(),b(C(e.as),{"aria-hidden":e.decorative?"true":void 0,class:"kui-icon check-icon","data-testid":"kui-icon-wrapper-check-icon",style:L(m.value)},{default:i(()=>[(c(),v("svg",{"aria-hidden":e.decorative?"true":void 0,"data-testid":"kui-icon-svg-check-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:_},null,8,H))]),_:1},8,["aria-hidden","style"]))}}),z=e=>(B("data-v-e0b0cf7b"),e=e(),M(),e),K=z(()=>s("strong",null,"few minutes",-1)),D={"data-testid":"kuma-environment"},F=z(()=>s("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),G={class:"item-status-list-wrapper"},U={class:"item-status-list"},Z={class:"circle mr-2"},j=I({__name:"OnboardingWelcomeView",setup(e){return(n,r)=>{const m=h("RouteTitle"),g=h("AppView"),f=h("RouteView");return c(),b(f,{name:"onboarding-welcome-view"},{default:i(({env:_,t:a,can:t})=>[d(m,{title:a("onboarding.routes.welcome.title",{name:a("common.product.name")}),render:!1},null,8,["title"]),o(),d(g,null,{default:i(()=>[s("div",null,[d(A,null,{header:i(()=>[d(E,null,{title:i(()=>[o(`
                Welcome to `+p(a("common.product.name")),1)]),description:i(()=>[s("p",null,[o(`
                  Congratulations on downloading `+p(a("common.product.name"))+"! You are just a ",1),K,o(` away from getting your service mesh fully online.
                `)]),o(),s("p",null,[o(`
                  We have automatically detected that you are running on `),s("strong",D,p(a(`common.product.environment.${_("KUMA_ENVIRONMENT")}`)),1),o(`.
                `)])]),_:2},1024)]),content:i(()=>[F,o(),s("div",G,[s("ul",U,[(c(!0),v(V,null,$([{name:`Run ${a("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...t("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],l=>(c(),v("li",{key:l.name},[s("span",Z,[l.status?(c(),b(S(W),{key:0,size:S(q)},null,8,["size"])):O("",!0)]),o(" "+p(l.name),1)]))),128))])])]),navigation:i(()=>[d(T,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}}),Y=R(j,[["__scopeId","data-v-e0b0cf7b"]]);export{Y as default};
